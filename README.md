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
陰影付きの六面、ホログラム状のゴーストライン、鼓動するカメラワークで近未来スクリーンセーバー風に仕上げています。

```bash
go run ./cmd/cybercube
```

### Neon Rain

奥行きレイヤーごとに速度と揺らぎが変わるデジタルレイン。  
列の着地にはスプラッシュが散り、手前のグローと奥の霧で立体感を演出します。

```bash
go run ./cmd/animterm -mode rain
```

### Spectrum Scope

ピークホールド付きの周波数バーと厚みのある走査波形を重ねたアナログ風スペクトラムアナライザー。  
横断するスキャンビームでVUメーター的なダイナミクスを追加しています。

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
