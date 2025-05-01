from aws_cdk import (
    aws_stepfunctions as sfn,
    aws_stepfunctions_tasks as tasks,
    aws_lambda as lambda_,
    Duration,
)
from constructs import Construct
import os

class StepFunctionsConstruct(Construct):
    """
    Creates the Step Functions workflow for the verification process.
    Implements the two-turn conversation workflow for image verification.
    """
    
    def __init__(
        self,
        scope: Construct,
        id: str,
        project_prefix: str,
        stage: str,
        lambda_functions,
        resource_suffix: str,
        **kwargs
    ) -> None:
        super().__init__(scope, id, **kwargs)
        self.resource_suffix = resource_suffix
        
        # ===================================
        # Create Lambda tasks for Step Functions
        # ===================================
        
        # Initialization tasks
        initialize_task = tasks.LambdaInvoke(
            self,
            "Initialize",
            lambda_function=lambda_functions["initialize"],
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.verificationContext"
        )
        
        fetch_historical_verification_task = tasks.LambdaInvoke(
            self,
            "FetchHistoricalVerification",
            lambda_function=lambda_functions["fetch_historical_verification"],
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.historicalContext"
        )
        
        fetch_images_task = tasks.LambdaInvoke(
            self,
            "FetchImages",
            lambda_function=lambda_functions["fetch_images"],
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$"
        )
        
        prepare_system_prompt_task = tasks.LambdaInvoke(
            self,
            "PrepareSystemPrompt",
            lambda_function=lambda_functions["prepare_system_prompt"],
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.systemPrompt"
        )
        
        # Initialize Conversation State (Pass state)
        initialize_conversation_state = sfn.Pass(
            self,
            "InitializeConversationState",
            parameters={
                "verificationContext.$": "$.verificationContext",
                "images.$": "$.images",
                "systemPrompt.$": "$.systemPrompt",
                "historicalContext.$": "$.historicalContext",
                "conversationState": {
                    "currentTurn": 0,
                    "maxTurns": 2,
                    "history": [],
                    "referenceAnalysis": {},
                    "checkingAnalysis": {}
                }
            }
        )
        
        # Turn 1 tasks (Reference image analysis)
        prepare_turn1_prompt_task = tasks.LambdaInvoke(
            self,
            "PrepareTurn1Prompt",
            lambda_function=lambda_functions["prepare_turn_prompt"],
            payload=sfn.TaskInput.from_object({
                "verificationContext.$": "$.verificationContext",
                "images.$": "$.images",
                "systemPrompt.$": "$.systemPrompt",
                "historicalContext.$": "$.historicalContext",
                "conversationState.$": "$.conversationState",
                "turnNumber": 1,
                "includeImage": "reference"
            }),
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.currentPrompt"
        )
        
        execute_turn1_task = tasks.LambdaInvoke(
            self,
            "ExecuteTurn1",
            lambda_function=lambda_functions["invoke_bedrock"],
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.turn1Response"
        )
        
        process_turn1_response_task = tasks.LambdaInvoke(
            self,
            "ProcessTurn1Response",
            lambda_function=lambda_functions["process_turn1_response"],
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.referenceAnalysis"
        )
        
        # Update conversation state after Turn 1 (Pass state)
        update_conversation_state_after_turn1 = sfn.Pass(
            self,
            "UpdateConversationStateAfterTurn1",
            parameters={
                "verificationContext.$": "$.verificationContext",
                "images.$": "$.images",
                "systemPrompt.$": "$.systemPrompt",
                "historicalContext.$": "$.historicalContext",
                "currentPrompt.$": "$.currentPrompt",
                "turn1Response.$": "$.turn1Response",
                "referenceAnalysis.$": "$.referenceAnalysis",
                "conversationState": {
                    "currentTurn": 1,
                    "maxTurns": 2,
                    "history.$": "States.Array($.conversationState.history, $.turn1Response)",
                    "referenceAnalysis.$": "$.referenceAnalysis",
                    "checkingAnalysis.$": "$.conversationState.checkingAnalysis"
                }
            }
        )
        
        # Turn 2 tasks (Checking image analysis & comparison)
        prepare_turn2_prompt_task = tasks.LambdaInvoke(
            self,
            "PrepareTurn2Prompt",
            lambda_function=lambda_functions["prepare_turn_prompt"],
            payload=sfn.TaskInput.from_object({
                "verificationContext.$": "$.verificationContext",
                "images.$": "$.images",
                "systemPrompt.$": "$.systemPrompt",
                "historicalContext.$": "$.historicalContext",
                "conversationState.$": "$.conversationState",
                "turnNumber": 2,
                "includeImage": "checking",
                "previousContext.$": "$.referenceAnalysis"
            }),
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.currentPrompt"
        )
        
        execute_turn2_task = tasks.LambdaInvoke(
            self,
            "ExecuteTurn2",
            lambda_function=lambda_functions["invoke_bedrock"],
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.turn2Response"
        )
        
        process_turn2_response_task = tasks.LambdaInvoke(
            self,
            "ProcessTurn2Response",
            lambda_function=lambda_functions["process_turn2_response"],
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.checkingAnalysis"
        )
        
        # Update conversation state after Turn 2 (Pass state)
        update_conversation_state_after_turn2 = sfn.Pass(
            self,
            "UpdateConversationStateAfterTurn2",
            parameters={
                "verificationContext.$": "$.verificationContext",
                "images.$": "$.images",
                "systemPrompt.$": "$.systemPrompt",
                "historicalContext.$": "$.historicalContext",
                "currentPrompt.$": "$.currentPrompt",
                "turn1Response.$": "$.turn1Response",
                "turn2Response.$": "$.turn2Response",
                "referenceAnalysis.$": "$.referenceAnalysis",
                "checkingAnalysis.$": "$.checkingAnalysis",
                "conversationState": {
                    "currentTurn": 2,
                    "maxTurns": 2,
                    "history.$": "States.Array($.conversationState.history, $.turn2Response)",
                    "referenceAnalysis.$": "$.referenceAnalysis",
                    "checkingAnalysis.$": "$.checkingAnalysis"
                }
            }
        )
        
        # Result finalization tasks
        finalize_results_task = tasks.LambdaInvoke(
            self,
            "FinalizeResults",
            lambda_function=lambda_functions["finalize_results"],
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.finalResults"
        )
        
        store_results_task = tasks.LambdaInvoke(
            self,
            "StoreResults",
            lambda_function=lambda_functions["store_results"],
            payload_response_only=True,
            retry_on_service_exceptions=True,
            result_path="$.storageResult"
        )
        
        notify_task = tasks.LambdaInvoke(
            self,
            "Notify",
            lambda_function=lambda_functions["notify"],
            payload_response_only=True,
            result_path="$.notificationResult"
        )
        
        # Final workflow state (Pass state)
        workflow_complete = sfn.Pass(
            self,
            "WorkflowComplete",
            parameters={
                "verificationId.$": "$.verificationContext.verificationId",
                "verificationType.$": "$.verificationContext.verificationType",
                "status": "COMPLETED",
                "timestamp.$": "$$.State.EnteredTime",
                "result": {
                    "verificationStatus.$": "$.finalResults.verificationStatus",
                    "resultImageUrl.$": "$.storageResult.resultImageUrl",
                    "confidenceScore.$": "$.finalResults.confidenceScore",
                    "discrepanciesCount.$": "$.finalResults.discrepanciesCount"
                }
            }
        )
        
        # Error handling states
        handle_initialization_error = sfn.Pass(
            self,
            "HandleInitializationError",
            result_path="$.error",
            parameters={
                "status": "FAILED",
                "error": "Failed to initialize verification process"
            }
        )
        
        handle_historical_fetch_error = sfn.Pass(
            self,
            "HandleHistoricalFetchError",
            result_path="$.error",
            parameters={
                "status": "FAILED",
                "error": "Failed to retrieve historical verification data"
            }
        )
        
        handle_fetch_images_error = sfn.Pass(
            self,
            "HandleFetchImagesError",
            result_path="$.error",
            parameters={
                "status": "FAILED",
                "error": "Failed to fetch images or metadata"
            }
        )        
        # Create separate error handling tasks for each turn
        handle_bedrock_error_turn1_task = tasks.LambdaInvoke(
            self,
            "HandleBedrockErrorTurn1",
            lambda_function=lambda_functions["handle_bedrock_error"],
            payload_response_only=True,
            result_path="$"
        )
        
        handle_bedrock_error_turn2_task = tasks.LambdaInvoke(
            self,
            "HandleBedrockErrorTurn2",
            lambda_function=lambda_functions["handle_bedrock_error"],
            payload_response_only=True,
            result_path="$"
        )
        
        # Create separate finalize error tasks for each turn
        finalize_with_error_turn1_task = tasks.LambdaInvoke(
            self,
            "FinalizeWithErrorTurn1",
            lambda_function=lambda_functions["finalize_with_error"],
            payload_response_only=True,
            result_path="$.finalResults"
        )
        
        finalize_with_error_turn2_task = tasks.LambdaInvoke(
            self,
            "FinalizeWithErrorTurn2",
            lambda_function=lambda_functions["finalize_with_error"],
            payload_response_only=True,
            result_path="$.finalResults"
        )
        
        # Choice states
        verification_type_choice = sfn.Choice(self, "CheckVerificationType")
        should_notify_choice = sfn.Choice(self, "ShouldNotify")
        
        # ===================================
        # Chain together the workflow
        # ===================================
        
        # 1. Initialize and check verification type
        initialize_chain = (
            initialize_task
            .add_catch(
                handle_initialization_error,
                errors=["States.ALL"],
                result_path="$.error"
            )
            .next(verification_type_choice)
        )
        
        # 2. Branch based on verification type
        verification_type_choice.when(
            sfn.Condition.string_equals("$.verificationContext.verificationType", "previous_vs_current"),
            fetch_historical_verification_task
            .add_catch(
                handle_historical_fetch_error,
                errors=["States.ALL"],
                result_path="$.error"
            )
        ).otherwise(fetch_images_task)
        
        # 3. Continue with common workflow
        fetch_historical_verification_task.next(fetch_images_task)
        
        # 4. Rest of the workflow
        image_processing_chain = (
            fetch_images_task
            .add_catch(
                handle_fetch_images_error, 
                errors=["States.ALL"],
                result_path="$.error"
            )
            .next(prepare_system_prompt_task)
            .next(initialize_conversation_state)
            .next(prepare_turn1_prompt_task)
            .next(execute_turn1_task
                .add_catch(
                    handle_bedrock_error_turn1_task.next(finalize_with_error_turn1_task).next(store_results_task),
                    errors=["States.ALL"],
                    result_path="$.error"
                )
            )
            .next(process_turn1_response_task)
            .next(update_conversation_state_after_turn1)
            .next(prepare_turn2_prompt_task)
            .next(execute_turn2_task
                .add_catch(
                    handle_bedrock_error_turn2_task.next(finalize_with_error_turn2_task).next(store_results_task),
                    errors=["States.ALL"],
                    result_path="$.error"
                )
            )
            .next(process_turn2_response_task)
            .next(update_conversation_state_after_turn2)
            .next(finalize_results_task)
            .next(store_results_task)
            .next(should_notify_choice)
        )
        
        # 5. Notification choice
        should_notify_choice.when(
            sfn.Condition.boolean_equals("$.verificationContext.notificationEnabled", True),
            notify_task.next(workflow_complete)
        ).otherwise(workflow_complete)
        
        # Note: The error handling for execute_turn1_task and execute_turn2_task
        # has been implemented in the workflow chain section above
        
        # Create the state machine from the workflow definition
        self.state_machine = sfn.StateMachine(
            self,
            f"{project_prefix}-verification-workflow-{self.resource_suffix}",
            state_machine_name=f"{project_prefix}-verification-workflow-{stage}-{self.resource_suffix}",
            definition=initialize_chain,
            timeout=Duration.minutes(15)
        )