// services_/events/notifier.go
package events

type GroupNotifier interface {
	NotifyMemberAdded(groupID string, userEmail string) error
}
