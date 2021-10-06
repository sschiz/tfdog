package repository

// Repository describes a storage
// to save user subscriptions.
type Repository interface {
	SaveSubscription(sub Subscription) error
	RemoveSubscription(sub Subscription) error
	GetUserSubscriptions(userID int) ([]Subscription, error)
	GetAllSubscriptions() ([]Subscription, error)
}

// Subscription describes user subscription.
type Subscription struct {
	UserID  int
	Link    string
	AppName string
}
