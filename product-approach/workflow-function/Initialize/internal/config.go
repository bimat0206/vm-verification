package internal

// Config holds the application configuration from environment variables
type Config struct {
	// DynamoDB table names
	LayoutTable       string
	VerificationTable string
	
	// Prefixes and naming
	VerificationPrefix string
	
	// S3 bucket names
	ReferenceBucket string
	CheckingBucket  string
	StateBucket     string // New bucket for S3 state management
	
	// Default TTL for DynamoDB items in days
	DefaultTTLDays int
}

// GetDefaultConfig returns a config with sensible defaults
func GetDefaultConfig() Config {
	return Config{
		VerificationPrefix: "verif-",
		DefaultTTLDays:     30, // 30 days default TTL
	}
}

// ConversationConfig defines configuration for the conversation
type ConversationConfig struct {
	Type     string
	MaxTurns int
}

// ProcessRequest represents a unified input request format for service processing
type ProcessRequest struct {
	// Schema version fields for standardized format
	SchemaVersion       string
	VerificationContext interface{}
	
	// Direct fields (legacy format)
	VerificationType      string
	ReferenceImageUrl     string
	CheckingImageUrl      string
	VendingMachineId      string
	LayoutId              int
	LayoutPrefix          string
	PreviousVerificationId string
	ConversationConfig    ConversationConfig
	RequestId             string
	RequestTimestamp      string
	NotificationEnabled   bool
}

// HistoricalContext represents data from previous verifications
type HistoricalContext struct {
	PreviousVerificationId     string               `json:"previousVerificationId"`
	PreviousVerificationAt     string               `json:"previousVerificationAt"`
	PreviousVerificationStatus string               `json:"previousVerificationStatus"`
	HoursSinceLastVerification float64              `json:"hoursSinceLastVerification"`
	MachineStructure           *MachineStructure    `json:"machineStructure,omitempty"`
	VerificationSummary        *VerificationSummary `json:"verificationSummary,omitempty"`
	CheckingStatus             map[string]string    `json:"checkingStatus,omitempty"`
}

// MachineStructure contains information about the vending machine layout
type MachineStructure struct {
	RowCount      int      `json:"rowCount"`
	ColumnsPerRow int      `json:"columnsPerRow"`
	RowOrder      []string `json:"rowOrder"`
	ColumnOrder   []string `json:"columnOrder"`
}

// VerificationSummary contains summary information from a previous verification
type VerificationSummary struct {
	TotalPositionsChecked  int     `json:"totalPositionsChecked"`
	CorrectPositions       int     `json:"correctPositions"`
	DiscrepantPositions    int     `json:"discrepantPositions"`
	MissingProducts        int     `json:"missingProducts"`
	IncorrectProductTypes  int     `json:"incorrectProductTypes"`
	UnexpectedProducts     int     `json:"unexpectedProducts"`
	EmptyPositionsCount    int     `json:"emptyPositionsCount"`
	OverallAccuracy        float64 `json:"overallAccuracy"`
	OverallConfidence      float64 `json:"overallConfidence"`
	VerificationStatus     string  `json:"verificationStatus"`
	VerificationOutcome    string  `json:"verificationOutcome"`
}