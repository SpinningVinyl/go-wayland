package client

// Destroy preserves the pre-1.26 local cleanup behavior.
func (i *Compositor) Destroy() error { i.Context().Unregister(i); return nil }

// Destroy preserves the pre-1.26 local cleanup behavior.
func (i *Shm) Destroy() error { i.Context().Unregister(i); return nil }

// Destroy preserves the pre-1.26 local cleanup behavior.
func (i *DataDeviceManager) Destroy() error { i.Context().Unregister(i); return nil }
