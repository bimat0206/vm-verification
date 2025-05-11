package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/your-org/your-project/internal"
)

func main() {
	// Set required environment variables
	os.Setenv("REFERENCE_BUCKET", "kootoro-dev-s3-reference-x1y2z3")
	os.Setenv("CHECKING_BUCKET", "kootoro-dev-s3-checking-f6d3xl")

	// Create a sample input with the issue
	input := &internal.Input{
		VerificationContext: &internal.VerificationContext{
			VerificationID:     "verif-20240511-123456",
			VerificationAt:     "2024-05-11T12:34:56Z",
			Status:             "INITIATED",
			VerificationType:   "LAYOUT_VS_CHECKING",
			VendingMachineID:   "VM-12345",
			LayoutID:           12345,
			LayoutPrefix:       "layout_",
			ReferenceImageURL:  "s3://kootoro-dev-s3-reference-x1y2z3/path/to/reference.jpg",
			CheckingImageURL:   "s3://wrong-bucket-name/path/to/checking.jpg", // This should cause the error
		},
		LayoutMetadata: &internal.LayoutMetadata{
			MachineStructure: &internal.MachineStructure{
				RowCount:      5,
				ColumnsPerRow: 8,
				RowOrder:      []string{"A", "B", "C", "D", "E"},
				ColumnOrder:   []string{"1", "2", "3", "4", "5", "6", "7", "8"},
			},
		},
	}

	// Print input for verification
	inputJSON, _ := json.MarshalIndent(input, "", "  ")
	fmt.Println("Input:")
	fmt.Println(string(inputJSON))

	// Validate the input
	err := internal.ValidateInput(input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Validation passed")
	}

	// Fix the input and try again
	input.VerificationContext.CheckingImageURL = "s3://kootoro-dev-s3-checking-f6d3xl/path/to/checking.jpg"
	
	// Validate again
	err = internal.ValidateInput(input)
	if err != nil {
		fmt.Printf("Error after fix: %v\n", err)
	} else {
		fmt.Println("Validation passed after fix")
	}
}