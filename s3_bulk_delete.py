import boto3

def delete_bucket_contents(s3, bucket_name):
    try:
        versioning_status = s3.get_bucket_versioning(Bucket=bucket_name)
    except Exception as e:
        print(f"Error getting versioning status for {bucket_name}: {e}")
        return

    if versioning_status.get('Status') == 'Enabled':
        paginator = s3.get_paginator('list_object_versions')
        for page in paginator.paginate(Bucket=bucket_name):
            versions = page.get('Versions', []) + page.get('DeleteMarkers', [])
            if versions:
                objects_to_delete = [{'Key': v['Key'], 'VersionId': v['VersionId']} for v in versions]
                try:
                    s3.delete_objects(Bucket=bucket_name, Delete={'Objects': objects_to_delete})
                except Exception as e:
                    print(f"Error deleting objects in {bucket_name}: {e}")
    else:
        paginator = s3.get_paginator('list_objects_v2')
        for page in paginator.paginate(Bucket=bucket_name):
            if 'Contents' in page:
                objects_to_delete = [{'Key': obj['Key']} for obj in page['Contents']]
                try:
                    s3.delete_objects(Bucket=bucket_name, Delete={'Objects': objects_to_delete})
                except Exception as e:
                    print(f"Error deleting objects in {bucket_name}: {e}")

def delete_bucket(s3, bucket_name):
    delete_bucket_contents(s3, bucket_name)
    try:
        s3.delete_bucket(Bucket=bucket_name)
        print(f"Bucket {bucket_name} deleted successfully.")
    except Exception as e:
        print(f"Failed to delete bucket {bucket_name}: {e}")

def print_table(buckets):
    # Determine column widths
    idx_width = len(str(len(buckets))) + 2
    name_width = max(len("Bucket Name"), *(len(b['Name']) for b in buckets)) + 2
    date_width = len("Creation Date") + 2

    # Header
    header = f"{'No.':<{idx_width}}{'Bucket Name':<{name_width}}{'Creation Date':<{date_width}}"
    print(header)
    print('-' * (idx_width + name_width + date_width))

    # Rows
    for idx, bucket in enumerate(buckets, 1):
        creation = bucket['CreationDate'].strftime("%Y-%m-%d %H:%M:%S")
        print(f"{idx:<{idx_width}}{bucket['Name']:<{name_width}}{creation:<{date_width}}")

def main():
    s3 = boto3.client('s3')
    prefix = input("Enter the prefix of the bucket names to search: ").strip()
    response = s3.list_buckets()
    buckets = response['Buckets']
    buckets_to_delete = [
        {
            'Name': bucket['Name'],
            'CreationDate': bucket['CreationDate']
        }
        for bucket in buckets if prefix in bucket['Name']
    ]

    if not buckets_to_delete:
        print("No buckets found with prefix:", prefix)
        return

    print("\nThe following buckets will be deleted:\n")
    print_table(buckets_to_delete)

    print("\n*Note: AWS does not store the 'created by' information in S3 bucket metadata. "
          "To find out who created a bucket, you must query CloudTrail logs if logging was enabled at the time of creation.*\n")

    confirm = input("Are you sure you want to delete these buckets? (yes/no): ")
    if confirm.lower() != 'yes':
        print("Aborted.")
        return

    for bucket in buckets_to_delete:
        print(f"Deleting bucket: {bucket['Name']}")
        delete_bucket(s3, bucket['Name'])

if __name__ == "__main__":
    main()
