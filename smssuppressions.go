package bird

// SmsSuppressionsService reads and edits who an SMS sender may not message.
// Reach it via Client.SmsSuppressions. A suppression covers one sender and one
// subscriber, so the same number can appear under several of them.
type SmsSuppressionsService struct{ resource }
