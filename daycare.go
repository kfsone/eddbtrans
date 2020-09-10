package eddbtrans

import (
	"sync"
)

// Utility for allowing out-of-order registration and lookup of parent<->child relations,
// for example querying whether the system id of a station is valid, without knowing
// whether the system will have been loaded yet.

type parentCheck struct {
	parentID uint32
	entity   interface{}
}

// Daycare is a place for parents to register and children to look them up.
type Daycare struct {
	requests  chan interface{}
	approvals chan interface{}
	openWG    sync.WaitGroup
	registry  map[uint32][]interface{}

	// statistics
	Registered uint64
	Queried    uint64
	Approved   uint64
	Queued     uint64
	Dequeued   uint64
	Duplicate  uint64
}

// Close releases memory used by the Daycare once the registrations and inquiries
// channels have been closed.
func (dc *Daycare) Close() (err error) {
	<- dc.approvals
	dc.registry = nil
	return nil
}

func (dc *Daycare) Registry() map[uint32][]interface{} {
	return dc.registry
}

// Approvals returns a channel from which approvals may be consumed.
func (dc *Daycare) Approvals() <-chan interface{} {
	return dc.approvals
}

// CloseInquiries closes down the inquiry channel.
func (dc *Daycare) CloseRegistration() {
	dc.openWG.Done()
}

func (dc *Daycare) CloseLookups() {
	dc.openWG.Done()
}

// Lookup forwards an entity to be matched against its parent by the Daycare center.
func (dc *Daycare) Lookup(parentID uint32, entity interface{}) {
	dc.requests <- parentCheck{parentID: parentID, entity: entity}
}

// Register adds an entity to the registry, approving any pending and future queries.
func (dc *Daycare) Register(entityID uint32) {
	dc.requests <- entityID
	dc.Registered++
}

func (dc *Daycare) register(id uint32) {
	// Register an id if it's not already registered.
	if waiting, existed := dc.registry[id]; existed == true {
		// Had children tried to look up this parent?
		if waiting != nil {
			for _, child := range waiting {
				dc.approvals <- child
				dc.Dequeued++
				dc.Approved++
			}
		} else {
			dc.Duplicate++
		}
	} else {
		dc.Registered++
	}

	// Prevent future children from registering
	dc.registry[id] = nil
}

func (dc *Daycare) lookup(check parentCheck) {
	// See if the parent exists, or register us as waiting for it
	// For the parent to be registered, it should exist but point to nil.
	if waiting, existed := dc.registry[check.parentID]; existed == false {
		dc.registry[check.parentID] = append(make([]interface{}, 0, 8), check.entity)
		dc.Queued++
	} else if waiting != nil {
		dc.registry[check.parentID] = append(waiting, check.entity)
		dc.Queued++
	} else {
		// registered, approve
		dc.approvals <- check.entity
		dc.Approved++
	}
	dc.Queried++
}

// OpenDayCare launches a worker which will monitor incoming registrations and inquiries,
// which will be processed through a registry. Inquiries for parents not yet registered are
// placed into a waiting queue. When registrations come in, they check if they have a queue
// and if so, immediately approve the waiters in the order they arrived.
func OpenDayCare() (dc *Daycare) {
	// Create the instance with channels.
	dc = &Daycare{
		requests:  make(chan interface{}, 4),
		approvals: make(chan interface{}, 1),
		registry:  make(map[uint32][]interface{}),
	}

	// Start the background worker.
	dc.openWG.Add(2) // to channels to wait on
	go func() {
		defer close(dc.requests)
		dc.openWG.Wait()
	}()

	go func() {
		defer close(dc.approvals)

		for received := range dc.requests {
			switch received.(type) {
			case uint32: // receiving a registration
				dc.register(received.(uint32))

			case parentCheck:
				dc.lookup(received.(parentCheck))
			}
		}

		orphans := make(map[uint32][]interface{}, 64)
		for id, list := range dc.registry {
			if list != nil && len(list) > 0 {
				orphans[id] = list
			}
		}
		dc.registry = orphans
	}()

	return
}
