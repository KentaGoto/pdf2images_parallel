package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

func runCommand(paths []string, imgResolution string) string {
	wg := &sync.WaitGroup{}

	flag := 0 // PDFファイルが有るかどうかのフラグ

	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			//log.Println(runtime.NumGoroutine()) // goroutineの数
			defer wg.Done()
			ext := strings.LastIndex(path, ".") // 拡張子（.pdf）
			// PDFが存在したら以下の処理をする
			if path[ext:] == ".pdf" {
				flag = 1
				pdfDir := filepath.Dir(path)            // PDFファイルのディレクトリパス
				filename := getFileNameWithoutExt(path) // ファイル名
				saveDir := pdfDir + "/" + filename      // 抽出画像を保存するフォルダ
				err1 := os.Mkdir(saveDir, 0777)         // PDFファイル名のフォルダを作る
				if err1 != nil {
					panic(err1)
				}
				// mutoolを叩いて画像出力
				imageFile := saveDir + "/" + "p%04d.png"
				cmd := exec.Command("mutool.exe", "draw", "-o", imageFile, "-r", imgResolution, path)
				err2 := cmd.Run()
				if err2 != nil {
					panic(err2)
				}
				fmt.Println("\n", path) // 処理後のファイルをフルパスで出力
				// 実行コマンドのステータスを表示
				state := cmd.ProcessState
				fmt.Printf("  %s\n", state.String())               // 終了コードと状態
				fmt.Printf("    Pid: %d\n", state.Pid())           // プロセスID
				fmt.Printf("    System: %v\n", state.SystemTime()) // システム時間（カーネル内で行われた処理の時間）
				fmt.Printf("    User: %v\n", state.UserTime())     // ユーザー時間（プロセス内で消費された時間）
			}
		}(path)
	}
	wg.Wait() // goroutineが終わるまで待つ

	// PDFが無ければ終了
	if flag == 0 {
		fmt.Println("PDF file is missing.")
		os.Exit(1)
	}

	return "\nDone."
}

func getFileNameWithoutExt(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

func main() {
	// 第1引数が-hか--helpだったらUsageなどを出力して終了
	if len(os.Args) == 2 {
		help := os.Args[1]
		if help == "-h" || help == "--help" {
			fmt.Println(`USAGE
  $> pdf2images_parallel.exe <DIR>

DESCRIPTION
  Convert each page of PDF to PNG.
  PNG is saved in a folder with the same name as PDF.
  It is recursively processed with the directory specified as the first argument as the root.

OPTION
  -h or --help

REQUIREMENTS
  Windows

AUTHOR
  KentaGoto`)
			os.Exit(1)
		}
	}

	// 引数が1つ以外は終了
	if len(os.Args) != 2 {
		fmt.Println("The number of arguments specified is incorrect. Only one argument is allowed.")
		os.Exit(1)
	}

	dir := os.Args[1]     // 第1引数（処理対象ルートディレクトリ）
	paths := dirwalk(dir) // ルートディレクトリを再帰で読みにいく

	// ファイルが何もなければ終了
	if paths == nil {
		fmt.Println("File is missing.")
		os.Exit(1)
	}

	fmt.Println("Processing...")

	// 画像解像度の設定
	const imgResolution = "300"

	// プログレスバー
	s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
	s.Color("green")
	s.Start()

	// コマンド起動
	result := runCommand(paths, imgResolution)

	s.Stop() // プログレスバー終了

	fmt.Println(result)
}
