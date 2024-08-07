// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package k8s

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/tools/cache"

	"github.com/cilium/hive/cell"
	"github.com/cilium/hive/job"
	"github.com/cilium/statedb"
	"github.com/cilium/stream"

	"github.com/cilium/cilium/pkg/time"
)

var errListerWatcherNil = errors.New("ReflectorConfig.ListerWatcher must be defined (was nil)")

// RegisterReflector registers a Kubernetes to StateDB table reflector.
//
// Intended to be used with [cell.Invoke] and the module's job group.
// See [ExampleRegisterReflector] for example usage.
func RegisterReflector[Obj any](jobGroup job.Group, db *statedb.DB, targetTable statedb.RWTable[Obj], cfg ReflectorConfig[Obj]) error {
	if cfg.ListerWatcher == nil {
		return errListerWatcherNil
	}

	// Register initializer that marks when the table has been initially populated,
	// e.g. the initial "List" has concluded.
	r := &k8sReflector[Obj]{
		ReflectorConfig: cfg.withDefaults(),
		db:              db,
		table:           targetTable,
	}
	wtxn := db.WriteTxn(targetTable)
	r.initDone = targetTable.RegisterInitializer(wtxn, "k8s-reflector")
	wtxn.Commit()

	jobGroup.Add(job.OneShot(
		fmt.Sprintf("k8s-reflector-[%T]", *new(Obj)),
		r.run))

	return nil
}

type ReflectorConfig[Obj any] struct {
	// Maximum number of objects to commit in one transaction. Uses default if left zero.
	// This does not apply to the initial listing which is committed in one go.
	BufferSize int

	// The amount of time to wait for the buffer to fill. Uses default if left zero.
	BufferWaitTime time.Duration

	// The ListerWatcher to use to retrieve the objects.
	//
	// Use [utils.ListerWatcherFromTyped] to create one from the Clientset, e.g.
	//
	//   var cs client.Clientset
	//   utils.ListerWatcherFromTyped(cs.CoreV1().Nodes())
	//
	ListerWatcher cache.ListerWatcher

	// Optional function to transform the objects given by the ListerWatcher. This can
	// be used to convert into an internal model on the fly to save space and add additional
	// fields or to for example implement TableRow/TableHeader for a cilium-dbg statedb command.
	Transform TransformFunc[Obj]

	// Optional function to query all objects. Used when replacing the objects on resync.
	// This can be used to "namespace" the objects managed by this reflector, e.g. on
	// source.Source etc.
	QueryAll QueryAllFunc[Obj]
}

// TransformFunc is an optional function to give to the Kubernetes reflector
// to transform the object returned by the ListerWatcher to the desired
// target object. If the function returns false the object is silently
// skipped.
type TransformFunc[Obj any] func(any) (obj Obj, ok bool)

// QueryAllFunc is an optional function to give to the Kubernetes reflector
// to query all objects in the table that are managed by the reflector.
// It is used to delete all objects when the underlying cache.Reflector needs
// to Replace() all items for a resync.
type QueryAllFunc[Obj any] func(statedb.ReadTxn, statedb.Table[Obj]) statedb.Iterator[Obj]

const (
	// DefaultBufferSize is the maximum number of objects to commit to the table in one write transaction.
	DefaultBufferSize = 10000

	// DefaultBufferWaitTime is the amount of time to wait to fill the buffer before committing objects.
	// 10000 * 50ms => 200k objects per second throughput limit.
	DefaultBufferWaitTime = 50 * time.Millisecond
)

// withDefaults fills in unset fields with default values.
func (cfg ReflectorConfig[Obj]) withDefaults() ReflectorConfig[Obj] {
	if cfg.BufferSize == 0 {
		cfg.BufferSize = DefaultBufferSize
	}
	if cfg.BufferWaitTime == 0 {
		cfg.BufferWaitTime = DefaultBufferWaitTime
	}
	return cfg
}

type k8sReflector[Obj any] struct {
	ReflectorConfig[Obj]

	initDone func(statedb.WriteTxn)
	db       *statedb.DB
	table    statedb.RWTable[Obj]
}

func (r *k8sReflector[Obj]) run(ctx context.Context, health cell.Health) error {
	type entry struct {
		deleted   bool
		name      string
		namespace string
		obj       Obj
	}
	type buffer struct {
		replaceItems []any
		entries      map[string]entry
	}
	bufferSize := r.BufferSize
	waitTime := r.BufferWaitTime
	table := r.table

	transform := r.Transform
	if transform == nil {
		// No provided transform function, use the identity function instead.
		transform = TransformFunc[Obj](func(obj any) (Obj, bool) { return obj.(Obj), true })
	}

	queryAll := r.QueryAll
	if queryAll == nil {
		// No query function provided, use All()
		queryAll = QueryAllFunc[Obj](func(txn statedb.ReadTxn, tbl statedb.Table[Obj]) statedb.Iterator[Obj] {
			return tbl.All(txn)
		})
	}

	// Construct a stream of K8s objects, buffered into chunks every [waitTime] period
	// and then committed.
	// This reduces the number of write transactions required and thus the number of times
	// readers get woken up, which results in much better overall throughput.
	src := stream.Buffer(
		ListerWatcherToObservable(r.ListerWatcher),
		bufferSize,
		waitTime,

		// Buffer the events into a map, coalescing them by key.
		func(buf *buffer, ev CacheStoreEvent) *buffer {
			switch {
			case ev.Kind == CacheStoreEventReplace:
				return &buffer{
					replaceItems: ev.Obj.([]any),
					entries:      make(map[string]entry, bufferSize), // Forget prior entries
				}
			case buf == nil:
				buf = &buffer{
					replaceItems: nil,
					entries:      make(map[string]entry, bufferSize),
				}
			}

			var entry entry
			var ok bool
			entry.obj, ok = transform(ev.Obj)
			if !ok {
				return buf
			}
			entry.deleted = ev.Kind == CacheStoreEventDelete

			meta, err := meta.Accessor(ev.Obj)
			if err != nil {
				panic(fmt.Sprintf("%T internal error: meta.Accessor failed: %s", r, err))
			}
			entry.name = meta.GetName()
			entry.namespace = meta.GetNamespace()
			var key string
			if entry.namespace != "" {
				key = entry.namespace + "/" + entry.name
			} else {
				key = entry.name
			}
			buf.entries[key] = entry
			return buf
		},
	)

	commitBuffer := func(buf *buffer) {
		numUpserted, numDeleted := 0, 0

		txn := r.db.WriteTxn(table)
		if buf.replaceItems != nil {
			iter := queryAll(txn, table)
			for obj, _, ok := iter.Next(); ok; obj, _, ok = iter.Next() {
				numDeleted++
				table.Delete(txn, obj)
			}
			for _, item := range buf.replaceItems {
				if obj, ok := transform(item); ok {
					table.Insert(txn, obj)
					numUpserted++
				}
			}
			// Mark the table as initialized. Internally this has a sync.Once
			// so safe to call multiple times.
			r.initDone(txn)
		}

		for _, entry := range buf.entries {
			if !entry.deleted {
				numUpserted++
				table.Insert(txn, entry.obj)
			} else {
				numDeleted++
				table.Delete(txn, entry.obj)
			}
		}

		numTotal := table.NumObjects(txn)
		txn.Commit()

		health.OK(fmt.Sprintf("%d inserted, %d deleted, %d total", numUpserted, numDeleted, numTotal))
	}

	errs := make(chan error)
	src.Observe(
		ctx,
		commitBuffer,
		func(err error) {
			errs <- err
			close(errs)
		},
	)
	return <-errs
}

// ListerWatcherToObservable turns a ListerWatcher into an observable using the
// client-go's Reflector.
func ListerWatcherToObservable(lw cache.ListerWatcher) stream.Observable[CacheStoreEvent] {
	return stream.FuncObservable[CacheStoreEvent](
		func(ctx context.Context, next func(CacheStoreEvent), complete func(err error)) {
			store := &cacheStoreListener{
				onAdd: func(obj any) {
					next(CacheStoreEvent{CacheStoreEventAdd, obj})
				},
				onUpdate:  func(obj any) { next(CacheStoreEvent{CacheStoreEventUpdate, obj}) },
				onDelete:  func(obj any) { next(CacheStoreEvent{CacheStoreEventDelete, obj}) },
				onReplace: func(objs []any) { next(CacheStoreEvent{CacheStoreEventReplace, objs}) },
			}
			reflector := cache.NewReflector(lw, nil, store, 0)
			go func() {
				reflector.Run(ctx.Done())
				complete(nil)
			}()
		})
}

type CacheStoreEventKind int

const (
	CacheStoreEventAdd = CacheStoreEventKind(iota)
	CacheStoreEventUpdate
	CacheStoreEventDelete
	CacheStoreEventReplace
)

type CacheStoreEvent struct {
	Kind CacheStoreEventKind
	Obj  any
}

// cacheStoreListener implements the methods used by the cache reflector and
// calls the given handlers for added, updated and deleted objects.
type cacheStoreListener struct {
	onAdd, onUpdate, onDelete func(any)
	onReplace                 func([]any)
}

func (s *cacheStoreListener) Add(obj interface{}) error {
	s.onAdd(obj)
	return nil
}

func (s *cacheStoreListener) Update(obj interface{}) error {
	s.onUpdate(obj)
	return nil
}

func (s *cacheStoreListener) Delete(obj interface{}) error {
	s.onDelete(obj)
	return nil
}

func (s *cacheStoreListener) Replace(items []interface{}, resourceVersion string) error {
	if items == nil {
		// Always emit a non-nil slice for replace.
		items = []interface{}{}
	}
	s.onReplace(items)
	return nil
}

// These methods are never called by cache.Reflector:

func (*cacheStoreListener) Get(obj interface{}) (item interface{}, exists bool, err error) {
	panic("unimplemented")
}
func (*cacheStoreListener) GetByKey(key string) (item interface{}, exists bool, err error) {
	panic("unimplemented")
}
func (*cacheStoreListener) List() []interface{} { panic("unimplemented") }
func (*cacheStoreListener) ListKeys() []string  { panic("unimplemented") }
func (*cacheStoreListener) Resync() error       { panic("unimplemented") }

var _ cache.Store = &cacheStoreListener{}
