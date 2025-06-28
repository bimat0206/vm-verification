import boto3
from botocore.exceptions import ClientError

# ==== CONFIG SECTION: UPDATE THESE VALUES ====
DYNAMODB_TABLES = [
    "kootoro-dev-dynamodb-conversation-history-f6d3xl",
    "kootoro-dev-dynamodb-verification-results-f6d3xl",
    # Add more table names as needed
]

# ==== END CONFIG SECTION ====

dynamodb_resource = boto3.resource("dynamodb")

def delete_all_items_with_progress(table_name):
    table = dynamodb_resource.Table(table_name)
    key_names = [k['AttributeName'] for k in table.key_schema]

    print(f"\nScanning DynamoDB table: {table_name}")
    items = []
    scan_kwargs = {
        'ProjectionExpression': ', '.join(f"#{k}" for k in key_names),
        'ExpressionAttributeNames': {f"#{k}": k for k in key_names}
    }
    done = False
    start_key = None
    while not done:
        if start_key:
            scan_kwargs['ExclusiveStartKey'] = start_key
        response = table.scan(**scan_kwargs)
        items.extend(response.get('Items', []))
        start_key = response.get('LastEvaluatedKey', None)
        done = start_key is None

    total = len(items)
    print(f"Total records found in '{table_name}': {total}")

    if total > 0:
        with table.batch_writer() as batch:
            for i, item in enumerate(items, 1):
                key = {k: item[k] for k in key_names}
                batch.delete_item(Key=key)
                print(f"  Deleted record [{i}/{total}]")
        print(f"Finished deleting all items in '{table_name}'.\n")
    else:
        print(f"No items to delete in '{table_name}'.\n")

if __name__ == "__main__":
    for table in DYNAMODB_TABLES:
        try:
            delete_all_items_with_progress(table)
        except Exception as e:
            print(f"Error cleaning DynamoDB table {table}: {e}")

    print("\nDynamoDB cleanup complete!")
