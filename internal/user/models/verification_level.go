package user

type VerificationLevel int

const (
	Banned        VerificationLevel = iota - 1 // -1 - banned
	Unverified    // 0 - unverified
	EmailVerified // 1
	AlphaVerified // 2
	BetaVerified  // 3
	PhoneVerified // 4
	IDVerified    // 5
	TrustedUser   // 6
)

func (v VerificationLevel) String() string {
	switch v {
	case Unverified:
		return "unverified"
	case EmailVerified:
		return "email_verified"
	case AlphaVerified:
		return "alpha_verified"
	case BetaVerified:
		return "beta_verified"
	case PhoneVerified:
		return "phone_verified"
	case IDVerified:
		return "id_verified"
	case TrustedUser:
		return "trusted_user"
	default:
		return "unknown"
	}
}
