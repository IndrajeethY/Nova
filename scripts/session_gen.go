package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           NovaUserbot Session Generator                      ║")
	fmt.Println("║                                                              ║")
	fmt.Println("║  This will generate a string session for your Telegram      ║")
	fmt.Println("║  account. Keep your session string safe and never share it! ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your API_ID (from https://my.telegram.org): ")
	apiIdStr, _ := reader.ReadString('\n')
	apiIdStr = strings.TrimSpace(apiIdStr)
	apiId, err := strconv.Atoi(apiIdStr)
	if err != nil {
		fmt.Println("Error: Invalid API_ID. Must be a number.")
		os.Exit(1)
	}

	fmt.Print("Enter your API_HASH: ")
	apiHash, _ := reader.ReadString('\n')
	apiHash = strings.TrimSpace(apiHash)
	if apiHash == "" {
		fmt.Println("Error: API_HASH cannot be empty.")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Connecting to Telegram...")

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:         int32(apiId),
		AppHash:       apiHash,
		SessionName:   "nova_session",
		MemorySession: true,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	if err := client.Connect(); err != nil {
		fmt.Printf("Error connecting: %v\n", err)
		os.Exit(1)
	}

	if err := client.Start(); err != nil {
		fmt.Printf("Error starting client: %v\n", err)
		os.Exit(1)
	}

	me, err := client.GetMe()
	if err != nil {
		fmt.Printf("Error getting user info: %v\n", err)
		os.Exit(1)
	}

	session := client.ExportSession()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Session Generated!                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Logged in as: %s %s (@%s)\n", me.FirstName, me.LastName, me.Username)
	fmt.Printf("User ID: %d\n", me.ID)
	fmt.Println()
	fmt.Println("Your STRING_SESSION:")
	fmt.Println("────────────────────────────────────────────────────────────────")
	fmt.Println(session)
	fmt.Println("────────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("⚠️  WARNING: Keep this session string SAFE and NEVER share it!")
	fmt.Println("    Anyone with this string can access your Telegram account.")
	fmt.Println()
	fmt.Println("Add this to your environment variables:")
	fmt.Printf("export STRING_SESSION=\"%s\"\n", session)
	fmt.Println()

	client.Stop()
}
