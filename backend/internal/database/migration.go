package database

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MigrationManager はデータベースマイグレーションを管理する
type MigrationManager struct {
	db            *DB
	migrationPath string
}

// NewMigrationManager は新しいマイグレーションマネージャーを作成する
func NewMigrationManager(db *DB, migrationPath string) *MigrationManager {
	return &MigrationManager{
		db:            db,
		migrationPath: migrationPath,
	}
}

// RunMigrations は全てのマイグレーションを実行する
func (mm *MigrationManager) RunMigrations() error {
	log.Println("データベースマイグレーションを開始します...")

	// マイグレーションファイルの取得
	files, err := mm.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("マイグレーションファイルの取得に失敗しました: %w", err)
	}

	// ファイルをソートして順番に実行
	sort.Strings(files)

	for _, file := range files {
		if err := mm.runMigrationFile(file); err != nil {
			return fmt.Errorf("マイグレーションファイル %s の実行に失敗しました: %w", file, err)
		}
		log.Printf("マイグレーションファイル %s を実行しました", file)
	}

	log.Println("全てのマイグレーションが正常に完了しました")
	return nil
}

// getMigrationFiles はマイグレーションファイルのリストを取得する
func (mm *MigrationManager) getMigrationFiles() ([]string, error) {
	var files []string

	err := filepath.WalkDir(mm.migrationPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// .sqlファイルのみを対象とし、init.sqlは除外
		if !d.IsDir() && strings.HasSuffix(path, ".sql") && !strings.HasSuffix(path, "init.sql") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// runMigrationFile は単一のマイグレーションファイルを実行する
func (mm *MigrationManager) runMigrationFile(filePath string) error {
	// ファイルを開く
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ファイルを開けませんでした: %w", err)
	}
	defer file.Close()

	// ファイル内容を読み取り、SQLステートメントに分割
	scanner := bufio.NewScanner(file)
	var sqlBuilder strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// コメント行をスキップ
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}

		sqlBuilder.WriteString(line)
		sqlBuilder.WriteString(" ")

		// セミコロンで終わる場合はSQLステートメントとして実行
		if strings.HasSuffix(line, ";") {
			sqlStatement := strings.TrimSpace(sqlBuilder.String())
			if sqlStatement != "" {
				if err := mm.executeSQL(sqlStatement); err != nil {
					return fmt.Errorf("SQLの実行に失敗しました: %s, エラー: %w", sqlStatement, err)
				}
			}
			sqlBuilder.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ファイルの読み取りに失敗しました: %w", err)
	}

	return nil
}

// executeSQL はSQLステートメントを実行する
func (mm *MigrationManager) executeSQL(sql string) error {
	_, err := mm.db.Exec(sql)
	return err
}

// CreateDatabase はデータベースを作成する（存在しない場合）
func (mm *MigrationManager) CreateDatabase(dbName string) error {
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_unicode_ci", dbName)
	_, err := mm.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("データベースの作成に失敗しました: %w", err)
	}
	
	log.Printf("データベース %s を作成しました（または既に存在します）", dbName)
	return nil
}

// DropDatabase はデータベースを削除する（テスト用）
func (mm *MigrationManager) DropDatabase(dbName string) error {
	sql := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err := mm.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("データベースの削除に失敗しました: %w", err)
	}
	
	log.Printf("データベース %s を削除しました", dbName)
	return nil
}