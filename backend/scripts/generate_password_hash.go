package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	var password string
	
	// コマンドライン引数からパスワードを取得
	if len(os.Args) >= 2 {
		password = os.Args[1]
	} else {
		// 引数がない場合は.envファイルから読み取り
		envPassword, err := readPasswordFromEnv()
		if err != nil {
			fmt.Println("使用方法:")
			fmt.Println("  go run generate_password_hash.go <password>")
			fmt.Println("  または .env ファイルに ADMIN_PASSWORD を設定してください")
			log.Fatal(err)
		}
		password = envPassword
		fmt.Printf(".envファイルからパスワードを読み取りました\n")
	}

	if password == "" {
		log.Fatal("パスワードが空です")
	}
	
	// bcryptでパスワードをハッシュ化
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("パスワードハッシュ化エラー: %v", err)
	}

	fmt.Printf("パスワード: %s\n", password)
	fmt.Printf("ハッシュ: %s\n", string(hashedBytes))
	
	// 検証
	err = bcrypt.CompareHashAndPassword(hashedBytes, []byte(password))
	if err != nil {
		log.Fatalf("検証エラー: %v", err)
	}
	
	fmt.Println("検証: OK")
}

// readPasswordFromEnv は.envファイルからADMIN_PASSWORDを読み取る
func readPasswordFromEnv() (string, error) {
	// プロジェクトルートの.envファイルを読み取り
	file, err := os.Open("../.env")
	if err != nil {
		return "", fmt.Errorf(".envファイルが見つかりません: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// コメント行や空行をスキップ
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// ADMIN_PASSWORD行を探す
		if strings.HasPrefix(line, "ADMIN_PASSWORD=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf(".envファイル読み取りエラー: %v", err)
	}
	
	return "", fmt.Errorf("ADMIN_PASSWORDが.envファイルに見つかりません")
}