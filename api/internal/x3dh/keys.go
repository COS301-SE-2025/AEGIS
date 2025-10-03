package x3dh

// IdentityKey represents a long-term identity key pair
type IdentityKey struct {
	UserID     string // UUID
	PublicKey  string // base64 or hex encoded
	PrivateKey string // Encrypted before storing (not used in bundle)
}

// SignedPreKey represents a medium-term prekey signed by the Identity Key
type SignedPreKey struct {
	UserID     string
	PublicKey  string
	PrivateKey string // Encrypted
	Signature  string // Signature of SPK.pub using IK.priv
	CreatedAt  string // Optional: useful for expiry/rotation
}

// OneTimePreKey represents a single-use key available for X3DH handshake
type OneTimePreKey struct {
	ID         string // Unique ID (e.g. UUID)
	UserID     string
	PublicKey  string
	PrivateKey string // Encrypted
	IsUsed     bool
	CreatedAt  string
}
