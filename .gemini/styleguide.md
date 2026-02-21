# Gemini Code Assist Review Style Guide

コードレビューを行う際は、コメントの視認性を高めるため、以下のルールに厳密に従ってください。

1. **バッジ（Badge）によるプレフィックスの義務化**
   すべてのレビューコメント（インラインコメントおよび全体サマリー）の先頭には、Shields.ioを利用した以下のMarkdownバッジを**必ず**挿入してください。指摘の重要度や性質に合わせて、最も適切なバッジを1つ選んでください。

   - 🚨 致命的なバグや修正必須の指摘（赤色）:
     `![MUST](https://img.shields.io/badge/MUST-red)`

   - 💡 強い推奨やパフォーマンス改善の提案（オレンジ色）:
     `![WANT](https://img.shields.io/badge/WANT-orange)`

   - 📝 軽微な指摘やタイポ、リファクタリング提案（黄色）:
     `![NITS](https://img.shields.io/badge/NITS-yellow)`

   - ✨ 素晴らしいコードや肯定的な評価（緑色）:
     `![GOOD](https://img.shields.io/badge/GOOD-success)`

   - ❓ 単なる情報共有や質問（青色）:
     `![INFO](https://img.shields.io/badge/INFO-blue)`

2. **ポジティブなコメントの分離（ノイズ削減）**
   インラインコメント（コード行に対するコメント）には、原則として `MUST`, `WANT`, `NITS` のいずれかのバッジを使用し、修正や改善が必要な箇所のみを指摘してください。
   `GOOD` バッジを使用するような称賛や肯定的なコメントは、インラインには書かず、Pull Request全体のサマリーコメント（全体レビュー）の中にまとめて記載してください。
