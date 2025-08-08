# S3 Diff Archive

A powerful, efficient command-line tool for incremental backup and archiving of files to Amazon S3. This tool performs differential backups by only archiving files that have changed since the last backup, making it ideal for large datasets where full backups would be inefficient.

## 🚀 Features

- **Incremental Backups**: Only archives files that have changed since the last backup
- **S3 Integration**: Direct upload to Amazon S3 with configurable storage classes
- **Password Protection**: Encrypt your archives with password-based encryption
- **File Filtering**: Support for exclude patterns using glob syntax
- **Multiple Tasks**: Configure multiple backup tasks in a single configuration file
- **Database Tracking**: Uses BadgerDB to track file states and changes
- **Compression**: Automatic ZIP compression with configurable size limits
- **Restoration**: Experimental restore functionality from archived backups (Must request deep_archive restore first from s3)
- **Detailed Logging**: Comprehensive logging for monitoring and debugging
- **Notifications**: Configurable notification system for operation status updates

## 📦 Installation

### Download Pre-built Binaries

Pre-compiled binaries are available for download from the [Releases](https://github.com/fahidsarker/s3-diff-archive/releases) section. Choose the appropriate binary for your operating system:

- **Linux**: `s3-diff-archive-linux-amd64`
- **macOS**: `s3-diff-archive-darwin-amd64` (Intel) or `s3-diff-archive-darwin-arm64` (Apple Silicon)
- **Windows**: `s3-diff-archive-windows-amd64.exe`

### Build from Source

If you prefer to build from source:

```bash
git clone https://github.com/fahidsarker/s3-diff-archive.git
cd s3-diff-archive
go build -o s3-diff-archive .
```

## ⚙️ Configuration

### Environment Variables

Create a `.env` file in your working directory with your AWS credentials:

```env
AWS_ACCESS_KEY_ID=your_access_key_here
AWS_SECRET_ACCESS_KEY=your_secret_key_here
AWS_REGION=us-east-1
S3_BUCKET=your-bucket-name
```

### Configuration File

Create a YAML configuration file (e.g., `config.yaml`) based on the sample:

```yaml
# Base path in S3 bucket where archives will be stored
s3_base_path: "backups/my-project"

# Directory to store logs (optional)
logs_dir: "./logs"

# Temporary directory for creating zip files
working_dir: "./tmp"

# Maximum size for each zip file in MB
max_zip_size: 5000

# Notification script for operation status updates (optional)
# Available placeholders: %icon%, %operation%, %status%, %message%
notify_script: 'echo "%icon% %operation% - %status% | %message%"'

# Backup tasks configuration
tasks:
  - id: photos
    dir: "./photos"
    storage_class: "DEEP_ARCHIVE"  # Cost-effective for long-term storage
    encryption_key: "MySecurePassword123"
    exclude: ["**/.DS_Store", "**/Thumbs.db", "**/*.tmp"]

  - id: documents
    dir: "./documents"
    storage_class: "STANDARD_IA"   # For infrequently accessed files
    encryption_key: "AnotherSecurePassword456"

  - id: videos
    dir: "./videos"
    storage_class: "GLACIER"       # Even more cost-effective for archives
```

### Storage Classes

Choose the appropriate S3 storage class based on your access patterns and cost requirements:

- **STANDARD**: For frequently accessed data
- **INTELLIGENT_TIERING**: Automatic cost optimization
- **STANDARD_IA**: For infrequently accessed data
- **ONEZONE_IA**: Lower cost for infrequently accessed data (single AZ)
- **GLACIER**: For archival data accessed once or twice per year
- **DEEP_ARCHIVE**: Lowest cost for long-term archival (7-10 years)

### Notification System

The tool supports configurable notifications for operation status updates. Configure the `notify_script` in your config file to receive notifications:

```yaml
# Simple echo notification (default)
notify_script: 'echo "%icon% %operation% - %status% | %message%"'

# macOS notification using osascript
notify_script: 'osascript -e "display notification \"%message%\" with title \"S3 Archive - %operation%\" subtitle \"%status%\""'

# Linux notification using notify-send
notify_script: 'notify-send "S3 Archive - %operation%" "%message%" --urgency=normal'

# Slack webhook notification
notify_script: 'curl -X POST -H "Content-type: application/json" --data "{\"text\":\"%icon% %operation% - %status%: %message%\"}" YOUR_SLACK_WEBHOOK_URL'

# Discord webhook notification
notify_script: 'curl -H "Content-Type: application/json" -d "{\"content\":\"%icon% %operation% - %status%: %message%\"}" YOUR_DISCORD_WEBHOOK_URL'
```

#### Available Placeholders

- `%icon%`: Status-specific emoji (✅ for success, ❌ for error, ⚠️ for warning, ❌⚠️❌⚠️ for fatal)
- `%operation%`: The operation being performed (scan, archive, restore, system)
- `%status%`: Operation status (success, error, warn, fatal)
- `%message%`: Detailed message about the operation result

## 🔧 Usage

### Basic Commands

```bash
# Scan directories for changes (dry run)
s3-diff-archive scan -config config.yaml

# Archive changed files to S3
s3-diff-archive archive -config config.yaml

# Restore files from S3
s3-diff-archive restore -config config.yaml

# View database contents for a specific task
s3-diff-archive view -config config.yaml -task photos
```

### Command-line Options

Each command supports the following flags:

- `-config`: Path to configuration file (required)
- `-env`: Path to environment file (default: `.env`)
- `-task`: Task ID (required for `view` command only)

### Example Workflow

1. **Initial Setup**:
   ```bash
   # Create your configuration
   cp config.sample.yaml config.yaml
   # Edit config.yaml with your settings
   
   # Set up environment variables
   cp .env.example .env
   # Edit .env with your actual AWS credentials
   ```

2. **Scan for Changes**:
   ```bash
   s3-diff-archive scan -config config.yaml
   ```

3. **Perform Backup**:
   ```bash
   s3-diff-archive archive -config config.yaml
   ```

4. **Restore When Needed**:
   ```bash
   s3-diff-archive restore -config config.yaml
   ```

## 📁 Project Structure

```
s3-diff-archive/
├── main.go                 # Main application entry point
├── go.mod                  # Go module dependencies
├── config.sample.yaml      # Sample configuration file
├── archiver/              
│   ├── archiver.go        # File archiving logic
│   └── zipper.go          # ZIP compression utilities
├── constants/
│   └── colors.go          # Terminal color constants
├── crypto/
│   ├── files.go           # File encryption/decryption
│   └── strings.go         # String encryption utilities
├── db/
│   ├── container.go       # Database container management
│   ├── db-archiver.go     # Database archiving
│   ├── db.go              # Main database operations
│   ├── reg.go             # File registry management
│   └── view.go            # Database viewing utilities
├── logger/
│   ├── log.go             # Logging configuration
│   └── loggers.go         # Logger implementations
├── restorer/
│   ├── compare.go         # File comparison utilities
│   └── restorer.go        # File restoration logic
├── s3/
│   ├── s3-manager.go      # S3 operations manager
│   └── task-uploader.go   # Task-specific upload logic
├── scanner/
│   ├── scanner.go         # File system scanning
│   └── types.go           # Scanner type definitions
├── types/
│   ├── s3-config.go       # S3 configuration types
│   └── sfile.go           # File metadata types
└── utils/
    ├── config-parser.go   # Configuration parsing
    ├── notifier.go        # Notification system
    ├── rand-create.go     # Random data generation
    ├── tools.go           # General utilities
    └── zipper.go          # ZIP file utilities
```

## 🔍 How It Works

1. **Scanning**: The tool scans specified directories and calculates checksums for all files
2. **Comparison**: File states are compared against a local BadgerDB database stored in S3
3. **Differential Detection**: Only files that have changed (new, modified, or deleted) are identified
4. **Archiving**: Changed files are compressed into password-protected ZIP archives
5. **Upload**: Archives are uploaded to S3 with the specified storage class
6. **Database Update**: The local database is updated and synchronized with S3

## 🛡️ Security Features

- **Encryption**: All archives are password-protected using ZIP encryption
- **AWS IAM**: Leverages AWS IAM for secure access control
- **Secure Storage**: Passwords are not stored in configuration files
- **Integrity Checking**: File checksums ensure data integrity

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. **Fork the Repository**
   ```bash
   git clone https://github.com/yourusername/s3-diff-archive.git
   cd s3-diff-archive
   ```

2. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make Your Changes**
   - Write clean, well-documented code
   - Follow Go best practices and conventions
   - Add tests for new functionality

4. **Test Your Changes**
   ```bash
   go test ./...
   go build .
   ```

5. **Submit a Pull Request**
   - Provide a clear description of your changes
   - Include any relevant issue numbers
   - Ensure all tests pass

### Development Guidelines

- **Code Style**: Follow standard Go formatting (`go fmt`)
- **Testing**: Add unit tests for new features
- **Documentation**: Update README and code comments as needed
- **Dependencies**: Minimize external dependencies when possible

### Reporting Issues

- Use the [GitHub Issues](https://github.com/fahidsarker/s3-diff-archive/issues) page
- Provide detailed reproduction steps
- Include configuration files (with sensitive data removed)
- Specify your operating system and Go version

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [BadgerDB](https://github.com/dgraph-io/badger) for efficient key-value storage
- [AWS SDK for Go](https://github.com/aws/aws-sdk-go-v2) for S3 integration
- [doublestar](https://github.com/bmatcuk/doublestar) for glob pattern matching

## 📞 Support

For support and questions:

- 📫 Create an issue on [GitHub Issues](https://github.com/fahidsarker/s3-diff-archive/issues)
- 📖 Check the documentation and examples above
- 🔍 Search existing issues for similar problems

---

**Note**: This tool is designed for efficient incremental backups. For initial backups of large datasets, the first run may take longer as it processes all files. Subsequent runs will be much faster as only changed files are processed.

This tool does not guarantee data integrity or security beyond the provided encryption and S3 storage features. Always test your backup and restore processes to ensure they meet your requirements.
