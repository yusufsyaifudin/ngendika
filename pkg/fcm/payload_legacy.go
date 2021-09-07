package fcm

// LegacyMessageNotification specifies the predefined, user-visible key-value pairs of the
// notification payload.
type LegacyMessageNotification struct {
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
	ChannelID    string `json:"android_channel_id,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Image        string `json:"image,omitempty"`
	Sound        string `json:"sound,omitempty"`
	Badge        string `json:"badge,omitempty"`
	Tag          string `json:"tag,omitempty"`
	Color        string `json:"color,omitempty"`
	ClickAction  string `json:"click_action,omitempty"`
	BodyLocKey   string `json:"body_loc_key,omitempty"`
	BodyLocArgs  string `json:"body_loc_args,omitempty"`
	TitleLocKey  string `json:"title_loc_key,omitempty"`
	TitleLocArgs string `json:"title_loc_args,omitempty"`
}

// LegacyMessage represents list of targets, options, and payload for HTTP JSON messages.
// COpy from: https://github.com/appleboy/go-fcm/blob/v0.1.5/message.go#L22-L60
// See https://firebase.google.com/docs/cloud-messaging/http-server-ref#send-downstream
// See https://firebase.google.com/docs/cloud-messaging/concept-options#notifications_and_data_messages
type LegacyMessage struct {
	To                       string                     `json:"to,omitempty"`
	RegistrationIDs          []string                   `json:"registration_ids,omitempty"`
	Condition                string                     `json:"condition,omitempty"`
	CollapseKey              string                     `json:"collapse_key,omitempty"`
	Priority                 string                     `json:"priority,omitempty"`
	ContentAvailable         bool                       `json:"content_available,omitempty"`
	MutableContent           bool                       `json:"mutable_content,omitempty"`
	DelayWhileIdle           bool                       `json:"delay_while_idle,omitempty"`
	TimeToLive               *uint                      `json:"time_to_live,omitempty"`
	DeliveryReceiptRequested bool                       `json:"delivery_receipt_requested,omitempty"`
	DryRun                   bool                       `json:"dry_run,omitempty"`
	RestrictedPackageName    string                     `json:"restricted_package_name,omitempty"`
	Notification             *LegacyMessageNotification `json:"notification,omitempty"`
	Data                     map[string]interface{}     `json:"data,omitempty"`
	Apns                     map[string]interface{}     `json:"apns,omitempty"`
	Webpush                  map[string]interface{}     `json:"webpush,omitempty"`
}

type LegacyResponse struct {
	MulticastID  int64                  `json:"multicast_id"`
	Success      int                    `json:"success"`
	Failure      int                    `json:"failure"`
	CanonicalIDs int                    `json:"canonical_ids,omitempty"`
	Results      []LegacyResponseResult `json:"results,omitempty"`

	// Device Group HTTP Response
	FailedRegistrationIDs []string `json:"failed_registration_ids,omitempty"`

	// Topic HTTP response
	MessageID int64 `json:"message_id,omitempty"`
	Error     error `json:"error,omitempty"`
}

// LegacyResponseResult ...
type LegacyResponseResult struct {
	MessageID      string `json:"message_id,omitempty"`
	RegistrationID string `json:"registration_id,omitempty"`
	Error          error  `json:"error,omitempty"`
}
