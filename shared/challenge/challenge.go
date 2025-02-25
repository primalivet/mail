package challenge

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func CalculateHMAC(password, challenge string) string {
	// Create a new HMAC by defining the hash type and the key
	h := hmac.New(md5.New, []byte(password))
	
	// Write challenge to the HMAC
	h.Write([]byte(challenge))
	
	// Get result and convert to hex string
	return hex.EncodeToString(h.Sum(nil))
}


func  Generate(serverHostname string) (string, string) {
	// Create a unique challenge string with timestamp and hostname
	timestamp := time.Now().Unix()
	random := time.Now().UnixNano() % 1000000
	challengeOriginal := fmt.Sprintf("<%d.%d@%s>", timestamp, random, serverHostname)
	
	challengeEncoded := base64.StdEncoding.EncodeToString([]byte(challengeOriginal))
	
	return challengeOriginal, challengeEncoded
}

func Process(username, password, encodedChallenge string) string {
	// Decode the base64 challenge from server
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedChallenge)
	if err != nil {
		return ""
	}
	challenge := string(decodedBytes)
	
	// Calculate HMAC-MD5 digest
	digest := CalculateHMAC(password, challenge)
	
	// Format the response as "username digest"
	response := fmt.Sprintf("%s %s", username, digest)
	
	// Encode response in base64
	return base64.StdEncoding.EncodeToString([]byte(response))
}

// TODO: Split in steps and have a first one to decode, then another to verify
func  Verify(accounts map[string]string, challenge, clientResponse string) bool {
	// Decode client's base64 response
	decodedBytes, err := base64.StdEncoding.DecodeString(clientResponse)
	if err != nil {
		return false
	}
	decodedResponse := string(decodedBytes)
	
	// Split response into username and digest
	parts := strings.SplitN(decodedResponse, " ", 2)
	if len(parts) != 2 {
		return false
	}
	username := parts[0]
	password, exists := accounts[username]
	if !exists {
		return false
	}
	clientDigest := parts[1]
	
	// Calculate expected digest using the same password and challenge
	expectedDigest := CalculateHMAC(password, challenge)
	
	// Compare digests
	return clientDigest == expectedDigest
}

