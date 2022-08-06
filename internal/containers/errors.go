package containers

// ErrNoContextID indicates an ID value within a context does not exist EG.
// containerID imageID etc.
type ErrNoContextID interface{ NoContextID() }
type errNoContextID struct{ error }

func (e errNoContextID) NoContextID() error { return e.error }
func IsErrNoContextID(err error) bool       { _, is := err.(ErrNoContextID); return is }
func noContextIDError(err error) error {
	if err == nil || IsErrNoContextID(err) {
		return err
	}
	return errNoContextID{err}
}

// ErrEngineOffline indicates the docker engine is offline
type ErrEngineOffline interface{ EngineOffline() }
type errEngineOffline struct{ error }

func (e errEngineOffline) EngineOffline() error { return e.error }
func IsErrEngineOffline(err error) bool         { _, is := err.(errEngineOffline); return is }
func engineOfflineError(err error) error {
	if err == nil || IsErrEngineOffline(err) {
		return err
	}
	return errEngineOffline{err}
}

// ErrNeedContainerRemove indicates the container needs to be removed or renamed EG. conflict
type ErrNeedContainerRemove interface{ NeedRemove() }
type errNeedContainerRemove struct{ error }

func (e errNeedContainerRemove) NeedRemove() error { return e.error }
func IsErrNeedContainerRemove(err error) bool      { _, is := err.(errNeedContainerRemove); return is }
func needContainerRemoveErr(err error) error {
	if err == nil || IsErrNeedContainerRemove(err) {
		return err
	}
	return errNeedContainerRemove{err}
}

// ErrNeedPortReallocation indicates the container's port needs to be realocated
type ErrNeedPortReallocation interface{ NeedReallocation() }
type errNeedPortReallocation struct{ error }

func (e errNeedContainerRemove) NeedReallocation() error { return e.error }
func IsErrNeedPortReallocation(err error) bool           { _, is := err.(errNeedPortReallocation); return is }
func needPortReallocation(err error) error {
	if err == nil || IsErrNeedPortReallocation(err) {
		return err
	}
	return errNeedPortReallocation{err}
}

// ErrNeedContainerReCreate indicates the container needs to be recreated
type ErrNeedContainerReCreate interface{ NeedReCreate() }
type errNeedContainerReCreate struct{ error }

func (e errNeedContainerRemove) NeedReCreate() error { return e.error }
func IsErrNeedContainerReCreate(err error) bool      { _, is := err.(errNeedContainerReCreate); return is }
func needContainerReCreate(err error) error {
	if err == nil || IsErrNeedContainerReCreate(err) {
		return err
	}
	return errNeedContainerReCreate{err}
}
