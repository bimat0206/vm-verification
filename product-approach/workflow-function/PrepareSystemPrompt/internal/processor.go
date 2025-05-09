package internal



// BuildTemplateData creates the data object for template rendering
func BuildTemplateData(input *Input) (TemplateData, error) {
	vCtx := input.VerificationContext
	data := TemplateData{
		VerificationType: vCtx.VerificationType,
		VerificationID:   vCtx.VerificationID,
		VerificationAt:   vCtx.VerificationAt,
		VendingMachineID: vCtx.VendingMachineID,
	}
	
	// Set machine structure based on verification type
	if vCtx.VerificationType == "LAYOUT_VS_CHECKING" && input.LayoutMetadata != nil {
		ms := input.LayoutMetadata.MachineStructure
		data.MachineStructure = ms
		data.RowCount = ms.RowCount
		data.ColumnCount = ms.ColumnsPerRow
		data.RowLabels = FormatArrayToString(ms.RowOrder)
		data.ColumnLabels = FormatArrayToString(ms.ColumnOrder)
		data.TotalPositions = ms.RowCount * ms.ColumnsPerRow
		
		// Extract product mappings if available
		if input.LayoutMetadata.ProductPositionMap != nil {
			data.ProductMappings = FormatProductMappings(input.LayoutMetadata.ProductPositionMap)
		}
		
		if input.LayoutMetadata.Location != "" {
			data.Location = input.LayoutMetadata.Location
		}
	} else if vCtx.VerificationType == "PREVIOUS_VS_CURRENT" && input.HistoricalContext != nil {
		hCtx := input.HistoricalContext
		data.PreviousVerificationID = hCtx.PreviousVerificationID
		data.PreviousVerificationAt = hCtx.PreviousVerificationAt
		data.PreviousVerificationStatus = hCtx.PreviousVerificationStatus
		data.HoursSinceLastVerification = hCtx.HoursSinceLastVerification
		
		// Set machine structure from historical context if available
		if hCtx.MachineStructure != nil {
			ms := hCtx.MachineStructure
			data.MachineStructure = ms
			data.RowCount = ms.RowCount
			data.ColumnCount = ms.ColumnsPerRow
			data.RowLabels = FormatArrayToString(ms.RowOrder)
			data.ColumnLabels = FormatArrayToString(ms.ColumnOrder)
			data.TotalPositions = ms.RowCount * ms.ColumnsPerRow
		}
		
		// Include previous verification summary if available
		if hCtx.VerificationSummary != nil {
			data.VerificationSummary = hCtx.VerificationSummary
		}
	}
	
	return data, nil
}

