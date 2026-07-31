package bird

// VerifyService is the Verify product namespace. Reach it via Client.Verify.
type VerifyService struct {
	// Verifications starts verifications and checks the passcodes recipients submit.
	Verifications *VerifyVerificationsService
}

// VerifyVerificationsService starts a verification, sending a one-time passcode,
// and checks the passcode a recipient submits.
type VerifyVerificationsService struct{ resource }
