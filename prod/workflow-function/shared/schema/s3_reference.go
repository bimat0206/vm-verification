package schema

// S3Reference represents a pointer to an object in S3.
type S3Reference struct {
	Bucket string `json:"bucket" dynamodbav:"bucket"`
	Key    string `json:"key" dynamodbav:"key"`
	Size   int64  `json:"size,omitempty" dynamodbav:"size,omitempty"`
}
