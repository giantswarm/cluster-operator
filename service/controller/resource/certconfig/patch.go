// NOTE this file is copied from operatorkit for migration purposes. The goal
// here is to get rid of the crud primitive and move to the implementation of
// the new handler interface eventually.
package certconfig

type patchType string

const (
	patchCreate patchType = "create"
	patchDelete patchType = "delete"
	patchUpdate patchType = "update"
)

// patch is a set of information required in order to reconcile the current
// state towards the desired state. patch is split into three parts: create,
// delete and update changes. The parts are passed as arguments to the CRUD
// Resource's ApplyCreateChange, ApplyDeleteChange and ApplyUpdateChange
// functions respectively. patch changes are guaranteed to be applied in that
// order (i.e. create, update, delete).
type patch struct {
	data map[patchType]interface{}
}

func newPatch() *patch {
	return &patch{
		data: make(map[patchType]interface{}, 3),
	}
}

func (p *patch) SetCreateChange(create interface{}) {
	p.data[patchCreate] = create
}

func (p *patch) SetDeleteChange(delete interface{}) {
	p.data[patchDelete] = delete
}

func (p *patch) SetUpdateChange(update interface{}) {
	p.data[patchUpdate] = update
}

func (p *patch) getCreateChange() (interface{}, bool) {
	create, ok := p.data[patchCreate]
	return create, ok
}

func (p *patch) getDeleteChange() (interface{}, bool) {
	delete, ok := p.data[patchDelete]
	return delete, ok
}

func (p *patch) getUpdateChange() (interface{}, bool) {
	update, ok := p.data[patchUpdate]
	return update, ok
}
