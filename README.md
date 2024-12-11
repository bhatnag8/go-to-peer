
# Go-To-Peer: A P2P File Sharing System in <img src="https://upload.wikimedia.org/wikipedia/commons/0/05/Go_Logo_Blue.svg" alt="Go Logo" width="70" />


## Overview
Go-To-Peer is a decentralized application inspired by BitTorrent that enables users to upload and download files seamlessly. Designed with Go’s robust concurrency model, this application handles file chunking, peer discovery, and concurrent file transfers with efficiency and reliability.

### Key Features
- **File Chunking**: Split and reconstruct files of any size. 
- **Peer Discovery**: Connect with peers for decentralized transfers.
- **Concurrent Transfers**: Maximize efficiency with parallel uploads/downloads.
- **Data Integrity**: Ensure secure and accurate file transfers with hash-based verification.
- **User-Friendly CLI**: Simple and intuitive command-line interface.

---

## Getting Started

### Prerequisites
- [Go](https://golang.org/dl/) 1.20+ installed
- A GitHub account for version control

### Installation
1. Clone the repository:
```
git clone https://github.com/bhatnag8/go-to-peer.git
cd go-to-peer
```

2. Install dependencies:
```
go mod tidy
```


---

## Usage

### Testing File Chunking and Reconstruction
Run the following command to test chunking and reconstruction with auto-generated files:
```
go run main.go -test-chunking
```

### Uploading Files
```
go run main.go -upload <file_path>
```

### Downloading Files
```
go run main.go -download <file_id>
```

---

## Roadmap
1. **File Chunking & Reconstruction**: ✅ Done
2. **Random Test Data Generation**: ✅ Done
3. **Peer Connectivity**: 🚧 In Progress
4. **Concurrent Transfers**: ⏳ Planned
5. **Polishing and Documentation**: ⏳ Planned

---

## Contributing
Contributions are welcome! If you have ideas or find bugs, open an issue or create a pull request.

---

## License
This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
