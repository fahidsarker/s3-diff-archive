# s3-diff-archive

`s3-diff-archive` is a tool designed to scan directories, archive updated files into zip files, and upload them to an S3 bucket. It supports excluding certain files, compressing into zip with optional encryption, and restoration from stored archives.

## Purpose

The tool automates the process of differential archiving by scanning directories for file changes, archiving only updated files, and uploading these archives to S3 for backup. This can efficiently manage backups by saving space and reducing upload times.

## Components

- **Scanner**: Scans directories for file changes.
- **Archiver**: Archives the updated files into zip files.
- **Uploader**: Uploads the archived zips to an S3 bucket.
- **Restorer**: Restores files from archived zips stored in S3.

## Configuration

The tool is configured using a YAML configuration file. Below is a sample configuration:

```yaml
aws_access_key_id: your_aws_access_key_id
aws_secret_access_key: your_aws_secret_access_key
aws_region: your_aws_region
s3_bucket: your_s3_bucket
s3_base_path: "your_s3_base_path"

logs_dir: "./logs"
working_dir: "./tmp"
max_zip_size: 5000 # in MB

tasks:
  - id: photos
    dir: "./path-to-photos"
    storage_class: "STANDARD"
    encryption_key: "your-encryption-password"
  - id: videos
    dir: "./path-to-videos"
    storage_class: "STANDARD"
```

## Usage
### Commands
- scan - Scans the specified directories for changed files.

- archive - Archives changed files into zips and uploads them to S3.

- restore (experimental) - Restores files from archived zip files stored in S3.

- view - Views archived data for a specific task.

### How to Run
1. Install the required dependencies.

2. Set Up Your Configuration: Modify the config.yaml file with your settings such as AWS credentials, directories to scan, and storage options.

3. Run the tool with the desired command:
```bash
$ s3-diff-archive <command> <config-file-path>
```
Replace <command> with one of the supported commands (scan, archive, restore, view) and <config-file-path> with the path to your configuration file.

Example:
```bash
$ s3-diff-archive scan ./config.yaml
```

## Experimental Features
- ExperimentalDownloadArchivedZips: Downloads archived zips from S3, useful for testing and ensures that restoration from S3 works as expected.

## Logging
Logs are stored in the directory specified by logs_dir in the configuration file. They provide detailed information on the scanning, archiving, and restoring processes.

## Contributions
Contributions are welcome! Please fork the repository and submit a pull request with your changes.

## License
This project is licensed under the MIT License - see the LICENSE file for details.

