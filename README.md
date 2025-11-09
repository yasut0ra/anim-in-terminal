# anim-in-terminal

ターミナルで動く幾何学・サイバーパンク系アニメーションの詰め合わせです。  
気分に合わせてモードを切り替え、スクリーンセーバー感覚で眺められます。

## 使い方

```bash
go run ./cmd/animterm -mode cybercube
```

`-mode` には `cybercube`, `rain`, `spectrum`, `cloud` を指定できます。  
オプション `-width`, `-height`, `-delay` で端末サイズやスピードを上書きできます。  
`cybercube` 時のみ `-cube-layout multi|single` で複数キューブと単一キューブを切り替えられます（デフォルト: `multi`）。

## アニメーション一覧

### Cyber Cube

立方体のワイヤーフレームが奥行きを保ちながら回転。  
陰影付きの六面とホログラム状ゴーストライン、脈動するカメラワークで近未来 HUD 風に仕上げています。  
最新バージョンではスケールや位置・回転速度が異なる複数のキューブを同時描画し、立体ディスプレイのようなレイヤー感を表現しています。  
昔ながらの単一キューブを眺めたい場合は `-cube-layout single` を指定してください。

```bash
go run ./cmd/cybercube
```

### Neon Rain

レイヤーごとに速度と揺らぎが異なるデジタルレイン。  
列の着地ではスプラッシュが散り、手前グローと奥の霧が立体感を演出します。

```bash
go run ./cmd/animterm -mode rain
```

### Spectrum Scope

ピークホールド付き周波数バーと厚みのある走査波形を重ねたアナログ風スペクトラムアナライザー。  
横断するスキャンビームで VU メーター的ダイナミクスをプラス。

```bash
go run ./cmd/animterm -mode spectrum
```

### Nebula Clouds

滑らかなノイズで生成した多層雲がゆっくり流れ、たまに稲光が走るシネマティックな空模様。

```bash
go run ./cmd/animterm -mode cloud
```

## ファイル構成

```
cmd/
  animterm/    # モード切り替えエントリーポイント
  cybercube/   # 旧キューブ単体エントリーポイント
internal/
  cloud/       # 雲エフェクト
  cybercube/   # ワイヤーフレームキューブ
  rain/        # デジタルレイン
  spectrum/    # スペクトラムアニメ
go.mod
README.md
```
