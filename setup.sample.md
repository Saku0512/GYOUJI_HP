# 3. ローカル環境のセットアップ(例)

## 環境

- `Windows11 Home`
- `WSL`
  - `Kali-Linux`

## ソフトウェアのダウンロード

- **`Docker & Docker Compose`**
  - [`Docker Desktop`](https://www.docker.com/products/docker-desktop/)にアクセス
  - `Download Docker Desktop`より`Download for Windows-AMD64`を選択
  - `Docker Desktop Installer.exe`を実行
  - `cmd`等でバージョンを確認

    ```bash
    # dockerのバージョン確認 (cmd)
    $ docker --version
    Docker version 28.4.0, build d8eb465 #  例
    
    # docker composeのバージョン確認 (cmd)
    $ docker compose version
    Docker Compose version v2.39.2-desktop.1 # 例
    ```

- **`Node.js`&`npm`**
  - [`Node.js`](https://nodejs.org/)にアクセス
  - `Node.js®を入手`より `Linux`用の`Node.js® v22.19.0(LTS)`と`npm`を`nvm`を使ってダウンロードする を選択
  - `Kali-Linx`で表示されるコードを実行する

    ```bash
    # curlの確認 (Kali-Linux)
    $ curl --version
    # curlがインストールされていない場合はaptなどからインストールする
    $ sudo apt update
    $ sudo apt install curl

    # nvmのダウンロードとインストール (Kali-Linux)
    $ curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash

    # 再起動の代わり (Kali-Linux)
    $ \. "$HOME/.nvm/nvm.sh"

    # Node.jsのダウンロードとインストール (Kali-Linux)
    nvm install 22

    # Node.jsのバージョン確認 (Kali-Linux)
    $ node -v
    v22.19.0 # 例

    # npmのバージョン確認 (Kali-Linux)
    $ npm -v
    10.9.3 # 例
    ```

- **`Go`**
  - [`Go`](https://golang.org/dl/)にアクセス
  - `Featured downloads`の`Microsoft Windows`の`go1.25.1.windows-amd64.msi`をダウンロードし、実行

    ```bash
    $ go version (cmd)
    go version go1.25.1 windows/amd64 # 例

    $ go version (Kali-Linux)
    go version go1.24.2 linux/amd64 # 例
    ```
