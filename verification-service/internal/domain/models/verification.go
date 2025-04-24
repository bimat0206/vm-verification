package models

import (
	"time"
)

// VerificationStatus represents the current state of the verification process
type VerificationStatus string

const (
	StatusInitialized      VerificationStatus = "INITIALIZED"
	StatusImagesFetched    VerificationStatus = "IMAGES_FETCHED"
	StatusSystemPromptReady VerificationStatus = "SYSTEM_PROMPT_READY"
	StatusTurn1PromptReady VerificationStatus = "TURN1_PROMPT_READY"
	StatusTurn1Processing  VerificationStatus = "TURN1_PROCESSING"
	StatusTurn1Completed   VerificationStatus = "TURN1_COMPLETED"
	StatusTurn2PromptReady VerificationStatus = "TURN2_PROMPT_READY"
	StatusTurn2Processing  VerificationStatus = "TURN2_PROCESSING"
	StatusTurn2Completed   VerificationStatus = "TURN2_COMPLETED"
	StatusResultsFinalized VerificationStatus = "RESULTS_FINALIZED"
	StatusResultsStored    VerificationStatus = "RESULTS_STORED"
	StatusNotificationSent VerificationStatus = "NOTIFICATION_SENT"
	StatusError            VerificationStatus = "ERROR"
	StatusPartialResults   VerificationStatus = "PARTIAL_RESULTS"
	StatusCorrect          VerificationStatus = "CORRECT"
	StatusIncorrect        VerificationStatus = "INCORRECT"
)

// DiscrepancyType represents the type of discrepancy found
type DiscrepancyType string

const (
	DiscrepancyIncorrectProductType DiscrepancyType = "Incorrect Product Type"
	DiscrepancyMissingProduct       DiscrepancyType = "Missing Product"
	DiscrepancyUnexpectedProduct    DiscrepancyType = "Unexpected Product"
	DiscrepancyIncorrectPosition    DiscrepancyType = "Incorrect Position"
	DiscrepancyIncorrectQuantity    DiscrepancyType = "Incorrect Quantity"
	DiscrepancyIncorrectOrientation DiscrepancyType = "Incorrect Orientation"
	DiscrepancyLabelNotVisible      DiscrepancyType = "Label Not Visible"
)

// DiscrepancySeverity represents the severity of a discrepancy
type DiscrepancySeverity string

const (
	SeverityLow    DiscrepancySeverity = "Low"
	SeverityMedium DiscrepancySeverity = "Medium"
	SeverityHigh   DiscrepancySeverity = "High"
)

// TurnConfig holds configuration for the verification turns
type TurnConfig struct {
	MaxTurns          int `json:"maxTurns"`
	ReferenceImageTurn int `json:"referenceImageTurn"`
	CheckingImageTurn int `json:"checkingImageTurn"`
	TurnTimeoutMs     int `json:"turnTimeoutMs"`
}

// VerificationContext contains the state and metadata for a verification process
type VerificationContext struct {
	VerificationID   string             `json:"verificationId"`
	VerificationAt   time.Time          `json:"verificationAt"`
	Status           VerificationStatus `json:"status"`
	VendingMachineID string             `json:"vendingMachineId"`
	LayoutID         int                `json:"layoutId"`
	LayoutPrefix     string             `json:"layoutPrefix"`
	ReferenceImageURL string             `json:"referenceImageUrl"`
	CheckingImageURL  string             `json:"checkingImageUrl"`
	TurnConfig       TurnConfig         `json:"turnConfig"`
	CurrentTurn      int                `json:"currentTurn"`
	TurnTimestamps   map[string]time.Time `json:"turnTimestamps"`
	NotificationEnabled bool              `json:"notificationEnabled"`
	ProcessingMetadata struct {
		RequestID  string    `json:"requestId"`
		StartTime  time.Time `json:"startTime"`
		RetryCount int       `json:"retryCount"`
		Timeout    int       `json:"timeout"`
	} `json:"processingMetadata"`
}

// MachineStructure represents the physical structure of a vending machine
type MachineStructure struct {
	RowCount      int      `json:"rowCount"`
	ColumnsPerRow int      `json:"columnsPerRow"`
	RowOrder      []string `json:"rowOrder"`
	ColumnOrder   []string `json:"columnOrder"`
	PhysicalOrientation struct {
		TopRow          string `json:"topRow"`
		LeftColumn      string `json:"leftColumn"`
		RowDirection    string `json:"rowDirection"` // "topToBottom" or "bottomToTop"
		ColumnDirection string `json:"columnDirection"` // "leftToRight" or "rightToLeft"
	} `json:"physicalOrientation"`
}

// ProductPosition represents a product at a specific position
type ProductPosition struct {
	Position   string  `json:"position"`
	Product    string  `json:"product"`
	IsPresent  bool    `json:"isPresent"`
	Quantity   *int    `json:"quantity,omitempty"`
	Confidence *int    `json:"confidence,omitempty"`
}

// Discrepancy represents a detected discrepancy between reference and checking images
type Discrepancy struct {
	Position          string            `json:"position"`
	Expected          string            `json:"expected"`
	Found             string            `json:"found"`
	Issue             DiscrepancyType   `json:"issue"`
	Confidence        int               `json:"confidence"`
	Evidence          string            `json:"evidence"`
	VerificationResult VerificationStatus `json:"verificationResult"`
	Severity          DiscrepancySeverity `json:"severity"`
	ImageCoordinates  *struct {
		X      int `json:"x"`
		Y      int `json:"y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"imageCoordinates,omitempty"`
}

// ReferenceAnalysis contains the results from Turn 1 (reference layout analysis)
type ReferenceAnalysis struct {
	TurnNumber         int                               `json:"turnNumber"`
	MachineStructure   MachineStructure                  `json:"machineStructure"`
	RowAnalysis        map[string]map[string]interface{} `json:"rowAnalysis"`
	ProductPositions   map[string]map[string]interface{} `json:"productPositions"`
	EmptyPositions     []string                          `json:"emptyPositions"`
	Confidence         int                               `json:"confidence"`
	InitialConfirmation string                           `json:"initialConfirmation"`
	OriginalResponse   string                            `json:"originalResponse"`
	CompletedAt        time.Time                         `json:"completedAt"`
}

// CheckingAnalysis contains the results from Turn 2 (checking image verification)
type CheckingAnalysis struct {
	TurnNumber        int                               `json:"turnNumber"`
	VerificationStatus VerificationStatus                `json:"verificationStatus"`
	Discrepancies     []Discrepancy                     `json:"discrepancies"`
	TotalDiscrepancies int                              `json:"totalDiscrepancies"`
	Severity          DiscrepancySeverity               `json:"severity"`
	RowAnalysis       map[string]map[string]interface{} `json:"rowAnalysis"`
	EmptySlotReport   struct {
		ReferenceEmptyRows      []string `json:"referenceEmptyRows"`
		CheckingEmptyRows       []string `json:"checkingEmptyRows"`
		CheckingPartiallyEmptyRows []string `json:"checkingPartiallyEmptyRows"`
		CheckingEmptyPositions  []string `json:"checkingEmptyPositions"`
		TotalEmpty              int      `json:"totalEmpty"`
	} `json:"emptySlotReport"`
	Confidence       int       `json:"confidence"`
	OriginalResponse string    `json:"originalResponse"`
	CompletedAt      time.Time `json:"completedAt"`
}

// VerificationResult represents the final verification result
type VerificationResult struct {
	VerificationID    string             `json:"verificationId"`
	VerificationAt    time.Time          `json:"verificationAt"`
	Status            VerificationStatus `json:"status"`
	VendingMachineID  string             `json:"vendingMachineId"`
	LayoutID          int                `json:"layoutId"`
	LayoutPrefix      string             `json:"layoutPrefix"`
	ReferenceImageURL string             `json:"referenceImageUrl"`
	CheckingImageURL  string             `json:"checkingImageUrl"`
	ResultImageURL    string             `json:"resultImageUrl"`
	MachineStructure  MachineStructure   `json:"machineStructure"`
	InitialConfirmation string            `json:"initialConfirmation"`
	CorrectedRows     []string           `json:"correctedRows"`
	EmptySlotReport   struct {
		ReferenceEmptyRows      []string `json:"referenceEmptyRows"`
		CheckingEmptyRows       []string `json:"checkingEmptyRows"`
		CheckingPartiallyEmptyRows []string `json:"checkingPartiallyEmptyRows"`
		CheckingEmptyPositions  []string `json:"checkingEmptyPositions"`
		TotalEmpty              int      `json:"totalEmpty"`
	} `json:"emptySlotReport"`
	ReferenceStatus   map[string]string `json:"referenceStatus"`
	CheckingStatus    map[string]string `json:"checkingStatus"`
	Discrepancies     []Discrepancy     `json:"discrepancies"`
	VerificationSummary struct {
		TotalPositionsChecked int                `json:"totalPositionsChecked"`
		CorrectPositions      int                `json:"correctPositions"`
		DiscrepantPositions   int                `json:"discrepantPositions"`
		MissingProducts       int                `json:"missingProducts"`
		IncorrectProductTypes int                `json:"incorrectProductTypes"`
		UnexpectedProducts    int                `json:"unexpectedProducts"`
		EmptyPositionsCount   int                `json:"emptyPositionsCount"`
		OverallAccuracy       float64            `json:"overallAccuracy"`
		OverallConfidence     int                `json:"overallConfidence"`
		VerificationStatus    VerificationStatus `json:"verificationStatus"`
		VerificationOutcome   string             `json:"verificationOutcome"`
	} `json:"verificationSummary"`
	Metadata struct {
		BedrockModel  string    `json:"bedrockModel"`
		CompletedAt   time.Time `json:"completedAt"`
		ProcessingTime int       `json:"processingTime"`
		TokenUsage struct {
			Input  int `json:"input"`
			Output int `json:"output"`
			Total  int `json:"total"`
		} `json:"tokenUsage"`
	} `json:"metadata"`
}