# modules/api_gateway/validators.tf

# API Gateway Request Validators
resource "aws_api_gateway_request_validator" "full_validator" {
  name                        = "${var.api_name}-full-validator"
  rest_api_id                 = aws_api_gateway_rest_api.api.id
  validate_request_body       = true
  validate_request_parameters = true
}

resource "aws_api_gateway_request_validator" "params_only_validator" {
  name                        = "${var.api_name}-params-validator"
  rest_api_id                 = aws_api_gateway_rest_api.api.id
  validate_request_body       = false
  validate_request_parameters = true
}