locals {
  test_function_arns = {
    initialize                    = "arn:aws:lambda:us-east-1:879654127886:function:kootoro-dev-lambda-initialize-f6d3xl"
    fetch_historical_verification = "arn:aws:lambda:us-east-1:879654127886:function:kootoro-dev-lambda-fetch-historical-f6d3xl"
    fetch_images                  = "arn:aws:lambda:us-east-1:879654127886:function:kootoro-dev-lambda-fetch-images-f6d3xl"
    prepare_system_prompt         = "arn:aws:lambda:us-east-1:879654127886:function:kootoro-dev-lambda-prepare-system-prompt-f6d3xl"
    execute_turn1_combined        = "arn:aws:lambda:us-east-1:879654127886:function:kootoro-dev-lambda-execute-turn1-combined-f6d3xl"
    execute_turn2_combined        = "arn:aws:lambda:us-east-1:879654127886:function:kootoro-dev-lambda-execute-turn2-combined-f6d3xl"
    finalize_results              = "arn:aws:lambda:us-east-1:879654127886:function:kootoro-dev-lambda-finalize-results-f6d3xl"

    finalize_with_error = "arn:aws:lambda:us-east-1:879654127886:function:kootoro-dev-lambda-finalize-with-error-f6d3xl"
  }
}

output "test_template" {
  value = templatefile("modules/step_functions/templates/state_machine_definition.tftpl", {
    function_arns               = local.test_function_arns
    region                      = "us-east-1"
    account_id                  = "879654127886"
    dynamodb_verification_table = "test-table"
    dynamodb_conversation_table = "test-conversation"
  })
}
