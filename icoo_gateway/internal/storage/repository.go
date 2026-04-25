package storage

type Reader[T any] interface {
	Get(id string) (T, bool)
	List() []T
}

type Creator[T any, I any] interface {
	Create(input I) (T, error)
}

type Updater[T any, I any] interface {
	Update(id string, input I) (T, error)
}

type Activator[T any] interface {
	Activate(id string) (T, error)
}

type Deactivator[T any] interface {
	Deactivate(id string) (T, error)
}

type Disabler[T any] interface {
	Disable(id string) (T, error)
}

type Heartbeater[T any] interface {
	Heartbeat(id string) (T, error)
}

type ListerByParent[T any] interface {
	ListByConversation(parentID string) []T
}

type MessageScopeReader[T any] interface {
	ListMessagesByScope(parentID, scope string) ([]T, bool)
}

type Appender[T any, I any] interface {
	AddMessage(parentID string, input I) (T, error)
}

type Setter[T any] interface {
	SetLastRunID(parentID, childID string) (T, error)
}

type MemberAdder[T any, I any] interface {
	AddMember(parentID string, input I) (T, error)
}

type MemberUpdater[T any, I any] interface {
	UpdateMember(parentID, childID string, input I) (T, error)
}

type MemberDeleter[T any] interface {
	DeleteMember(parentID, childID string) (T, error)
}

type MemberReader[T any] interface {
	ListMembers(parentID string) ([]T, bool)
}

type Recorder[T any, I any] interface {
	Record(input I) T
}
