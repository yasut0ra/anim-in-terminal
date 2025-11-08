# anim-in-terminal

ターミナルで動く幾何学サイバーパンク系アニメーション集です。

## 使い方

```bash
go run ./cmd/animterm -mode cybercube
```

`-mode` には `cybercube`, `rain`, `spectrum` を指定できます。 `-width`, `-height`, `-delay` で簡易調整も可能です。

## アニメーション一覧

### Cyber Cube

立方体のワイヤーフレームが奥行きを保ちながら回転します。  
陰影付きの六面とグローする頂点で近未来スクリーンセーバー風に仕上げています。

```bash
go run ./cmd/cybercube
```

### Neon Rain

デジタルな縦雨が列ごとのスピードで流れ落ちる「マトリクス」系エフェクト。

```bash
go run ./cmd/animterm -mode rain
```

### Spectrum Scope

周波数バーと走査波形を組み合わせたアナログ風スペクトラムアナライザー。

```bash
go run ./cmd/animterm -mode spectrum
```

## ファイル構成

```
cmd/
  animterm/    # モード切り替えエントリーポイント
  cybercube/   # 旧キューブ専用エントリーポイント
internal/
  cybercube/   # キューブ描画ロジック
  rain/        # デジタルレイン
  spectrum/    # スペクトラムアニメ
go.mod
README.md
```
