import boto3
import uuid
import json
import random
import time
import os
from datetime import datetime, timedelta
import base64
from decimal import Decimal
from boto3.dynamodb.conditions import Key
import argparse

# Command-line arguments
parser = argparse.ArgumentParser(description='Generate mock data for Kootoro vending machine verification Use Case 2')
parser.add_argument('--local', action='store_true', help='Use local DynamoDB endpoint')
parser.add_argument('--endpoint-url', type=str, help='Custom DynamoDB endpoint URL')
parser.add_argument('--region', type=str, help='AWS region (defaults to environment or profile)')
parser.add_argument('--records', type=int, default=10, help='Number of historical verification records to generate')
parser.add_argument('--vendor-machines', type=int, default=5, help='Number of vending machines to simulate')
parser.add_argument('--verification-table', type=str, help='Name of verification results table')
parser.add_argument('--conversation-table', type=str, help='Name of conversation history table')
parser.add_argument('--verbose', action='store_true', help='Enable verbose output')
parser.add_argument('--dry-run', action='store_true', help='Generate data but don\'t write to DynamoDB')
args = parser.parse_args()

# Configuration for DynamoDB
dynamodb_kwargs = {}
if args.region:
    dynamodb_kwargs['region_name'] = args.region

if args.local:
    dynamodb_kwargs['endpoint_url'] = 'http://localhost:8000'
    print("Using local DynamoDB at http://localhost:8000")
elif args.endpoint_url:
    dynamodb_kwargs['endpoint_url'] = args.endpoint_url
    print(f"Using custom DynamoDB endpoint at {args.endpoint_url}")

dynamodb = boto3.resource('dynamodb', **dynamodb_kwargs)
print(f"AWS Region: {dynamodb.meta.client.meta.region_name}")

# Table names - prioritize command line args, then env vars, then defaults
VERIFICATION_TABLE = args.verification_table or os.environ.get('VERIFICATION_TABLE', 'VerificationResults')
CONVERSATION_TABLE = args.conversation_table or os.environ.get('CONVERSATION_TABLE', 'ConversationHistory')

print(f"Using verification table: {VERIFICATION_TABLE}")
print(f"Using conversation table: {CONVERSATION_TABLE}")

# Verify tables exist
try:
    verification_table = dynamodb.Table(VERIFICATION_TABLE)
    verification_table.table_status  # This will trigger an error if table doesn't exist
    conversation_table = dynamodb.Table(CONVERSATION_TABLE)
    conversation_table.table_status
    print("Successfully connected to DynamoDB tables")
except Exception as e:
    print(f"Error connecting to tables: {str(e)}")
    print("Please ensure the tables exist and you have proper permissions.")
    exit(1)

# Helper functions
def generate_verification_id():
    """Generate a verification ID in the format 'verif-{timestamp}'"""
    timestamp = datetime.now().strftime('%Y%m%d%H%M%S')
    random_suffix = f"{random.randint(0, 99):02d}"
    return f"verif-{timestamp}{random_suffix}"

def generate_timestamp(days_ago=0, hours_ago=0, minutes_ago=0):
    """Generate an ISO 8601 timestamp for a time in the past"""
    dt = datetime.now() - timedelta(days=days_ago, hours=hours_ago, minutes=minutes_ago)
    return dt.strftime('%Y-%m-%dT%H:%M:%SZ')

def get_random_location():
    """Generate a random location"""
    buildings = ["Office Building A", "Mall Center B", "Transportation Hub C", "Hospital D", "University E"]
    floors = ["Ground Floor", "Floor 1", "Floor 2", "Floor 3", "Floor 4", "Floor 5"]
    return f"{random.choice(buildings)}, {random.choice(floors)}"

def get_random_vending_machine_id(num_machines=args.vendor_machines):
    """Generate a random vending machine ID"""
    return f"VM-{random.randint(1000, 1000 + num_machines - 1)}"

def get_machine_structure():
    """Get the standard machine structure"""
    return {
        "rowCount": 6,
        "columnsPerRow": 10,
        "rowOrder": ["A", "B", "C", "D", "E", "F"],
        "columnOrder": ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10"]
    }

def get_sample_products():
    """Get a list of sample products"""
    return [
        {"productId": 3486, "productName": "Mì Hảo Hảo", "color": "pink"},
        {"productId": 3920, "productName": "Mì Cung Đình", "color": "green"},
        {"productId": 4010, "productName": "Mi modern Lẩu thái", "color": "red/white"},
        {"productId": 4011, "productName": "Red Bull", "color": "blue/silver"},
        {"productId": 4012, "productName": "Coca Cola", "color": "red"},
        {"productId": 4013, "productName": "Pepsi", "color": "blue"}
    ]

def generate_reference_status(machine_structure):
    """Generate the reference status for each row"""
    products = get_sample_products()
    reference_status = {}
    
    for row in machine_structure["rowOrder"]:
        # Pick a random product for this row
        product = random.choice(products)
        reference_status[row] = f"Baseline: 7 {product['color']} '{product['productName']}' cup noodles visible."
    
    return reference_status

def generate_checking_status(reference_status, random_changes=True):
    """Generate the checking status based on reference, with possible changes"""
    checking_status = {}
    products = get_sample_products()
    
    for row, status in reference_status.items():
        try:
            # Try to extract product details safely
            parts = status.split("'")
            if len(parts) >= 3:
                product_name = parts[1]
                color_parts = status.split(" ")
                if len(color_parts) >= 2:
                    product_color = color_parts[1]
                else:
                    # Fallback if color not found
                    product_color = "unknown"
            else:
                # Fallback if product name not found in expected format
                product_name = f"Product-{row}"
                product_color = "unknown"
            
            # Decide if we'll make a change (20% chance)
            if random_changes and random.random() < 0.2:
                # Select a different product
                new_product = random.choice([p for p in products if p["productName"] != product_name])
                checking_status[row] = f"Current: 7 **{new_product['color']} '{new_product['productName']}'** cup noodles visible. Status: Changed Product."
            elif random_changes and random.random() < 0.1:
                # Make it empty
                checking_status[row] = f"Current: Empty row. Status: Products Removed."
            else:
                # No change
                checking_status[row] = f"Current: 7 {product_color} '{product_name}' cup noodles visible. Status: No Change."
        except Exception as e:
            # Fallback in case of any parsing issues
            print(f"Warning: Error parsing reference status for row {row}: {str(e)}")
            print(f"Status string was: {status}")
            random_product = random.choice(products)
            checking_status[row] = f"Current: 7 {random_product['color']} '{random_product['productName']}' cup noodles visible. Status: No Change."
    
    return checking_status

def generate_discrepancies(reference_status, checking_status, machine_structure):
    """Generate discrepancies based on the differences between reference and checking"""
    discrepancies = []
    
    for row in machine_structure["rowOrder"]:
        try:
            ref_status = reference_status[row]
            check_status = checking_status[row]
            
            # If there's a change, create discrepancy records
            if "No Change" not in check_status:
                # Extract product name safely
                ref_parts = ref_status.split("'")
                if len(ref_parts) >= 3:
                    ref_product = ref_parts[1]
                else:
                    ref_product = f"Unknown-Product-{row}"
                
                # Handle empty row case
                if "Empty row" in check_status:
                    for col in range(1, 8):  # Only create for visible columns 1-7
                        position = f"{row}{int(col):02d}"
                        discrepancies.append({
                            "position": position,
                            "expected": f"'{ref_product}' cup noodle",
                            "found": "Empty (Coils Visible)",
                            "issue": "Missing Product",
                            "confidence": random.randint(90, 99),
                            "evidence": "Coils clearly visible with no products",
                            "verificationResult": "INCORRECT"
                        })
                else:
                    # Product change case - try to extract the new product
                    check_parts = check_status.split("'")
                    if len(check_parts) >= 3:
                        check_product = check_parts[1]
                        # Try to extract color
                        if "**" in check_status:
                            color_parts = check_status.split("**")
                            if len(color_parts) >= 2:
                                check_color = color_parts[1].split(" ")[0]
                            else:
                                check_color = "different"
                        else:
                            check_color = "different"
                    else:
                        check_product = f"New-Product-{row}"
                        check_color = "different"
                    
                    for col in range(1, 8):  # Only create for visible columns 1-7
                        position = f"{row}{int(col):02d}"
                        discrepancies.append({
                            "position": position,
                            "expected": f"'{ref_product}' cup noodle",
                            "found": f"{check_color} '{check_product}' cup noodle",
                            "issue": "Incorrect Product Type",
                            "confidence": random.randint(90, 99),
                            "evidence": "Different packaging color and branding visible",
                            "verificationResult": "INCORRECT"
                        })
        except Exception as e:
            print(f"Warning: Error generating discrepancies for row {row}: {str(e)}")
            continue
    
    return discrepancies

def generate_verification_summary(discrepancies, machine_structure):
    """Generate the verification summary based on discrepancies"""
    total_positions = 42  # Visible positions (6 rows × 7 columns)
    discrepant_positions = len(discrepancies)
    correct_positions = total_positions - discrepant_positions
    
    # Count by issue type
    missing_products = sum(1 for d in discrepancies if d["issue"] == "Missing Product")
    incorrect_types = sum(1 for d in discrepancies if d["issue"] == "Incorrect Product Type")
    unexpected_products = sum(1 for d in discrepancies if d["issue"] == "Unexpected Product")
    
    # Calculate accuracy
    accuracy = round((correct_positions / total_positions) * 100, 1)
    
    empty_positions_count = missing_products
    
    return {
        "totalPositionsChecked": total_positions,
        "correctPositions": correct_positions,
        "discrepantPositions": discrepant_positions,
        "missingProducts": missing_products,
        "incorrectProductTypes": incorrect_types,
        "unexpectedProducts": unexpected_products,
        "emptyPositionsCount": empty_positions_count,
        "overallAccuracy": accuracy,
        "overallConfidence": random.randint(90, 98),
        "verificationStatus": "INCORRECT" if discrepant_positions > 0 else "CORRECT",
        "verificationOutcome": (
            "Discrepancies Detected - "
            f"{missing_products} missing products, "
            f"{incorrect_types} incorrect product types"
        ) if discrepant_positions > 0 else "Layout Verified - All Positions Match"
    }

def generate_empty_slot_report(discrepancies, machine_structure):
    """Generate the empty slot report based on discrepancies"""
    # Find rows with all positions empty
    empty_positions = [d["position"] for d in discrepancies if d["issue"] == "Missing Product"]
    empty_rows = []
    partially_empty_rows = {}
    
    for row in machine_structure["rowOrder"]:
        row_positions = [pos for pos in empty_positions if pos.startswith(row)]
        if len(row_positions) == 7:  # All visible positions in row are empty
            empty_rows.append(row)
        elif len(row_positions) > 0:  # Some positions in row are empty
            partially_empty_rows[row] = len(row_positions)
    
    return {
        "referenceEmptyRows": [],
        "checkingEmptyRows": empty_rows,
        "checkingPartiallyEmptyRows": [f"{row} ({count} empty)" for row, count in partially_empty_rows.items()],
        "checkingEmptyPositions": empty_positions,
        "totalEmpty": len(empty_positions)
    }

def generate_s3_url(bucket, path, filename):
    """Generate a mock S3 URL"""
    return f"s3://{bucket}/{path}/{filename}"

def generate_checking_image_urls(vending_machine_id, count=10, bucket="kootoro-checking-bucket"):
    """Generate a series of checking image URLs for a vending machine"""
    urls = []
    
    # Use different date formats for generating realistic S3 paths
    date_formats = [
        '%Y-%m-%d',  # Standard: 2025-05-01
        '%Y/%m/%d',  # Nested folders: 2025/05/01
        '%Y%m%d'     # Compact: 20250501
    ]
    
    # Randomly select a date format pattern for this machine
    date_format = random.choice(date_formats)
    
    # For newer images, use more recent dates - oldest image is 'count' days ago
    for i in range(count):
        # Generate timestamp with some random variation
        hours_variation = random.randint(0, 23)
        minutes_variation = random.randint(0, 59)
        
        timestamp = datetime.now() - timedelta(
            days=i, 
            hours=hours_variation,
            minutes=minutes_variation
        )
        
        # Format the date path according to the chosen pattern
        date_path = timestamp.strftime(date_format)
        
        # Format time with different possible patterns
        time_str = timestamp.strftime('%H-%M-%S')
        
        # Generate different filename patterns
        filename_patterns = [
            f"check_{time_str}.jpg",
            f"check_{timestamp.strftime('%Y%m%d_%H%M%S')}.jpg",
            f"{vending_machine_id}_{timestamp.strftime('%Y%m%d_%H%M%S')}.jpg",
            f"img_{uuid.uuid4().hex[:8]}.jpg"
        ]
        filename = random.choice(filename_patterns)
        
        # Generate the S3 URL
        if random.random() > 0.7:  # 30% chance of a different path structure
            url = f"s3://{bucket}/{vending_machine_id}/{date_path}/{filename}"
        else:
            url = f"s3://{bucket}/{date_path}/{vending_machine_id}/{filename}"
        
        urls.append((url, timestamp.strftime('%Y-%m-%dT%H:%M:%SZ')))
    
    return urls

def generate_conversation_history(verification_id, verification_timestamp, reference_img_url, checking_img_url, vendingMachineId, machine_structure):
    """Generate a mock conversation history for the verification"""
    # Calculate turn timestamps
    turn1_timestamp = (datetime.strptime(verification_timestamp, '%Y-%m-%dT%H:%M:%SZ') + timedelta(seconds=5)).strftime('%Y-%m-%dT%H:%M:%SZ')
    turn2_timestamp = (datetime.strptime(turn1_timestamp, '%Y-%m-%dT%H:%M:%SZ') + timedelta(seconds=15)).strftime('%Y-%m-%dT%H:%M:%SZ')
    
    return {
        "verificationId": verification_id,
        "conversationAt": verification_timestamp,
        "vendingMachineId": vendingMachineId,
        "currentTurn": 2,
        "maxTurns": 2,
        "turnStatus": "COMPLETED",
        "history": [
            {
                "turnId": 1,
                "timestamp": turn1_timestamp,
                "prompt": f"The FIRST image provided ALWAYS depicts the Previous State of the vending machine. This image shows how the vending machine {vendingMachineId} appeared during the last verification check.",
                "imageUrl": {
                    "reference": reference_img_url
                },
                "response": "I've analyzed the reference image showing the previous state of the vending machine. The machine has 6 rows (A-F) with products visible in each row...",
                "latencyMs": random.randint(1800, 2500),
                "tokenUsage": {
                    "input": random.randint(4000, 5000),
                    "output": random.randint(1500, 2000),
                    "total": random.randint(5500, 7000)
                },
                "analysisStage": "REFERENCE_ANALYSIS"
            },
            {
                "turnId": 2,
                "timestamp": turn2_timestamp,
                "prompt": f"The SECOND image provided ALWAYS depicts the Current State of the vending machine {vendingMachineId}. Compare this with the previous state to identify any changes or discrepancies.",
                "imageUrl": {
                    "checking": checking_img_url
                },
                "response": "After comparing the current state with the previous state, I've identified several changes...",
                "latencyMs": random.randint(1800, 2500),
                "tokenUsage": {
                    "input": random.randint(4500, 5500),
                    "output": random.randint(2000, 2500),
                    "total": random.randint(6500, 8000)
                },
                "analysisStage": "CHECKING_ANALYSIS"
            }
        ],
        "expiresAt": int((datetime.now() + timedelta(days=90)).timestamp()),
        "metadata": {
            "bedrockModel": "anthropic.claude-3-7-sonnet-20250219-v1:0",
            "systemPromptVersion": "1.2",
            "clientInfo": "Web/1.0.0"
        }
    }

def check_existing_tables():
    """Check that existing tables have the expected schema"""
    # This function just ensures the tables exist and have the minimum attributes we need
    # No table creation performed - we use existing tables only
    
    try:
        # Check VerificationResults table
        verification_table = dynamodb.Table(VERIFICATION_TABLE)
        table_desc = verification_table.meta.client.describe_table(TableName=VERIFICATION_TABLE)
        
        # Check primary key structure
        key_schema = table_desc['Table']['KeySchema']
        hash_key = next((k for k in key_schema if k['KeyType'] == 'HASH'), None)
        range_key = next((k for k in key_schema if k['KeyType'] == 'RANGE'), None)
        
        if not hash_key or hash_key['AttributeName'] != 'verificationId':
            print(f"Warning: {VERIFICATION_TABLE} does not have 'verificationId' as hash key")
            
        if not range_key or range_key['AttributeName'] != 'verificationAt':
            print(f"Warning: {VERIFICATION_TABLE} does not have 'verificationAt' as range key")
            
        # Check for required GSIs
        gsis = table_desc['Table'].get('GlobalSecondaryIndexes', [])
        gsi_names = [gsi['IndexName'] for gsi in gsis]
        
        required_indexes = ['CheckingImageIndex', 'ReferenceImageIndex']
        for idx in required_indexes:
            if idx not in gsi_names:
                print(f"Warning: {VERIFICATION_TABLE} does not have the {idx} GSI")
                print(f"  This might cause issues with UC2 query patterns")
                print(f"  Available GSIs: {', '.join(gsi_names)}")
        
        # Check ConversationHistory table
        conversation_table = dynamodb.Table(CONVERSATION_TABLE)
        table_desc = conversation_table.meta.client.describe_table(TableName=CONVERSATION_TABLE)
        
        # Check primary key structure
        key_schema = table_desc['Table']['KeySchema']
        hash_key = next((k for k in key_schema if k['KeyType'] == 'HASH'), None)
        range_key = next((k for k in key_schema if k['KeyType'] == 'RANGE'), None)
        
        if not hash_key or hash_key['AttributeName'] != 'verificationId':
            print(f"Warning: {CONVERSATION_TABLE} does not have 'verificationId' as hash key")
            
        if not range_key or range_key['AttributeName'] != 'conversationAt':
            print(f"Warning: {CONVERSATION_TABLE} does not have 'conversationAt' as range key")
            
        print("Table schema validation completed")
        
    except Exception as e:
        print(f"Error checking table schema: {str(e)}")
        print("Please ensure the tables exist and you have the proper permissions")
        exit(1)

def generate_use_case_2_data(num_records=args.records):
    """Generate mock data for use case 2 (previous vs current)"""
    print(f"Generating {num_records} records per machine for {args.vendor_machines} machines...")
    
    verification_table = dynamodb.Table(VERIFICATION_TABLE)
    conversation_table = dynamodb.Table(CONVERSATION_TABLE)
    
    verification_records_created = 0
    conversation_records_created = 0
    errors = 0
    
    s3_checking_bucket = os.environ.get("S3_CHECKING_BUCKET", "kootoro-checking-bucket")
    
    # Create records for each vending machine
    for vm_index in range(args.vendor_machines):
        vending_machine_id = get_random_vending_machine_id()
        location = get_random_location()
        machine_structure = get_machine_structure()
        
        print(f"Generating data for machine {vm_index+1}/{args.vendor_machines}: {vending_machine_id} at {location}")
        
        # Generate a series of checking image URLs for this machine
        image_urls = generate_checking_image_urls(vending_machine_id, count=num_records, bucket=s3_checking_bucket)
        
        # The first verification will be a standalone one (not linked to previous)
        reference_status = generate_reference_status(machine_structure)
        checking_status = generate_checking_status(reference_status)
        
        verification_ids = []
        
        # For each pair of consecutive images, create a verification
        for i in range(1, len(image_urls)):
            previous_img_url, previous_timestamp = image_urls[i]
            current_img_url, current_timestamp = image_urls[i-1]
            
            # Generate verification ID
            verification_id = generate_verification_id()
            verification_ids.append(verification_id)
            
            # Calculate hours since last verification
            prev_dt = datetime.strptime(previous_timestamp, '%Y-%m-%dT%H:%M:%SZ')
            curr_dt = datetime.strptime(current_timestamp, '%Y-%m-%dT%H:%M:%SZ')
            hours_diff = (curr_dt - prev_dt).total_seconds() / 3600
            
            # For each verification except the first, reference the previous one
            previous_verification_id = verification_ids[-2] if len(verification_ids) > 1 else None
            
            # Generate discrepancies between previous and current checking states
            discrepancies = generate_discrepancies(reference_status, checking_status, machine_structure)
            
            # Generate verification summary
            verification_summary = generate_verification_summary(discrepancies, machine_structure)
            
            # Generate empty slot report
            empty_slot_report = generate_empty_slot_report(discrepancies, machine_structure)
            
            # Create verification item
            verification_item = {
                "verificationId": verification_id,
                "verificationAt": current_timestamp,
                "verificationType": "PREVIOUS_VS_CURRENT",
                "vendingMachineId": vending_machine_id,
                "location": location,
                "referenceImageUrl": previous_img_url,
                "checkingImageUrl": current_img_url,
                "verificationStatus": verification_summary["verificationStatus"],
                "machineStructure": machine_structure,
                "initialConfirmation": f"Baseline analysis complete based on previous checking image. Row structure confirmed.",
                "correctedRows": [row for row, status in checking_status.items() if "No Change" in status],
                "emptySlotReport": empty_slot_report,
                "referenceStatus": reference_status,
                "checkingStatus": checking_status,
                "discrepancies": discrepancies,
                "verificationSummary": verification_summary,
                "metadata": {
                    "bedrockModel": "anthropic.claude-3-7-sonnet-20250219-v1:0",
                    "turn1LatencyMs": random.randint(1800, 2500),
                    "turn2LatencyMs": random.randint(1800, 2500),
                    "generatedBy": "mock-data-generator",
                    "generatedAt": datetime.now().strftime('%Y-%m-%dT%H:%M:%SZ')
                }
            }
            
            # Add previous verification ID if available
            if previous_verification_id:
                verification_item["previousVerificationId"] = previous_verification_id
            
            # Add historical context
            if previous_verification_id:
                verification_item["historicalContext"] = {
                    "previousVerificationId": previous_verification_id,
                    "previousVerificationAt": previous_timestamp,
                    "previousVerificationStatus": verification_summary["verificationStatus"],
                    "hoursSinceLastVerification": round(hours_diff, 1),
                    "machineStructure": machine_structure,
                    "verificationSummary": verification_summary
                }
            
            # Generate conversation history
            conversation_item = generate_conversation_history(
                verification_id, 
                current_timestamp, 
                previous_img_url, 
                current_img_url,
                vending_machine_id,
                machine_structure
            )
            
            # Convert to DynamoDB format (handling Decimal types)
            verification_item_db = json.loads(json.dumps(verification_item), parse_float=Decimal)
            conversation_item_db = json.loads(json.dumps(conversation_item), parse_float=Decimal)
            
            # Insert into DynamoDB (unless dry run)
            if not args.dry_run:
                try:
                    if args.verbose:
                        print(f"  Inserting verification record: {verification_id}")
                    
                    # Insert verification record
                    verification_table.put_item(Item=verification_item_db)
                    verification_records_created += 1
                    
                    # Insert conversation record
                    conversation_table.put_item(Item=conversation_item_db)
                    conversation_records_created += 1
                    
                except Exception as e:
                    errors += 1
                    print(f"Error inserting record: {str(e)}")
                    if args.verbose:
                        print(f"Verification item: {verification_id}")
                    continue
            else:
                if args.verbose:
                    print(f"  [DRY RUN] Would insert verification: {verification_id}")
                verification_records_created += 1
                conversation_records_created += 1
            
            # For the next iteration, use current checking status as the new reference status
            reference_status = checking_status
            checking_status = generate_checking_status(reference_status)
    
    # Print summary
    print("\nData generation summary:")
    print(f"  Verification records created: {verification_records_created}")
    print(f"  Conversation records created: {conversation_records_created}")
    if errors > 0:
        print(f"  Errors encountered: {errors}")
    
    if args.dry_run:
        print("\nDRY RUN: No records were actually inserted into DynamoDB")
    else:
        print("\nData generation complete!")
        print(f"Records inserted into:")
        print(f"  - {VERIFICATION_TABLE}")
        print(f"  - {CONVERSATION_TABLE}")
    
def main():
    """Main function"""
    # Check that tables exist and have suitable schema
    check_existing_tables()
    
    # Show configuration summary
    print("\nConfiguration:")
    print(f"  Records per machine: {args.records}")
    print(f"  Number of machines: {args.vendor_machines}")
    print(f"  Total records: ~{args.records * args.vendor_machines}")
    print(f"  Verification table: {VERIFICATION_TABLE}")
    print(f"  Conversation table: {CONVERSATION_TABLE}")
    if args.dry_run:
        print(f"  Dry run: Yes (no records will be inserted)")
    print()
    
    # Confirm with user if generating lots of records
    if args.records * args.vendor_machines > 100 and not args.dry_run:
        confirm = input(f"You're about to generate approximately {args.records * args.vendor_machines} records. Continue? (y/n): ")
        if confirm.lower() != 'y':
            print("Operation cancelled.")
            return
    
    # Generate data
    start_time = time.time()
    generate_use_case_2_data()
    end_time = time.time()
    
    # Print execution time
    print(f"\nExecution time: {end_time - start_time:.2f} seconds")

if __name__ == "__main__":
    main()