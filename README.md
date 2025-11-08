# anim-in-terminal

ターミナルで動く幾何学サイバーパンク系アニメーション集です。

## Cyber Cube

立方体のワイヤーフレームが回転し、背景にはバイナリのスキャンラインが流れます。  
スクリーンセーバー代わりに眺めてください。

```bash
go run ./cmd/cybercube
```

## ファイル構成

```
cmd/
  cybercube/   # エントリーポイント
internal/
  cybercube/   # 描画・投影ロジック
go.mod
README.md
```
