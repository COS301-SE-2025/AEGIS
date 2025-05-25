package main

import (
	"fmt"
	"aegis-api/services/evidence"
)

func main() {
	evidence.InitIPFSClient()

	cid, err := evidence.UploadToIPFS("hello.txt")
	if err != nil {
		fmt.Println("❌ Upload failed:", err)
		return
	}

	fmt.Println("✅ IPFS CID:", cid)
}
