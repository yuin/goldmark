# 基本的な書き込みと書式設定の構文

単純な構文を使用して、GitHubで散文とコードの高度な書式設定を作成します。

## 見出し

見出しを作成するには、見出しテキストの前に 1 \~ 6 個の <kbd>#</kbd> 記号を追加します。 使用する <kbd>#</kbd> の数によって、見出しの階層レベルと書体のサイズが決まります。

```markdown
# A first-level heading
## A second-level heading
### A third-level heading
```

![サンプル h1、h2、h3 ヘッダーを示す、レンダリングされたGitHub Markdown のスクリーンショット。階層レベルを示すために、型のサイズと視覚的な重みが下がります。](/assets/images/help/writing/headings-rendered.png)

2 つ以上の見出しを使用すると、GitHubファイル ヘッダー内の \[アウトライン] メニュー アイコン<svg version="1.1" width="16" height="16" viewBox="0 0 16 16" class="octicon octicon-list-unordered" aria-label="Table of Contents" role="img"><path d="M5.75 2.5h8.5a.75.75 0 0 1 0 1.5h-8.5a.75.75 0 0 1 0-1.5Zm0 5h8.5a.75.75 0 0 1 0 1.5h-8.5a.75.75 0 0 1 0-1.5Zm0 5h8.5a.75.75 0 0 1 0 1.5h-8.5a.75.75 0 0 1 0-1.5ZM2 14a1 1 0 1 1 0-2 1 1 0 0 1 0 2Zm1-6a1 1 0 1 1-2 0 1 1 0 0 1 2 0ZM2 4a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>をクリックしてアクセスできる目次が自動的に生成されます。 各見出しのタイトルは目次に表示され、タイトルをクリックして選択したセクションに移動できます。

![公開されている目次のドロップダウン メニューを含む README ファイルのスクリーンショット。 目次アイコンは濃いオレンジで囲まれています。](/assets/images/help/repository/headings-toc.png)

## テキストのスタイル設定

コメント フィールドや `.md` ファイルでは、太字、斜体、取り消し線、下付き文字、または上付き文字のテキストで強調を示すことができます。

| Style                                                                                  | 構文                                                | キーボード ショートカット                            | Example                                 | アウトプット                       |
| -------------------------------------------------------------------------------------- | ------------------------------------------------- | ---------------------------------------- | --------------------------------------- | ---------------------------- |
| 太字                                                                                     |                                                   |                                          |                                         |                              |
| `** **` または `__ __`                                                                    |                                                   |                                          |                                         |                              |
| <kbd>Command</kbd>+<kbd>B</kbd> (Mac) または <kbd>Ctrl</kbd>+<kbd>B</kbd> (Windows/Linux) | `**This is bold text**`                           |                                          |                                         |                              |
| **これは太字のテキストです**                                                                       |                                                   |                                          |                                         |                              |
| 斜体                                                                                     |                                                   |                                          |                                         |                              |
| `* *` または `_ _`                                                                        |                                                   |                                          |                                         |                              |
| <kbd>Command</kbd>+<kbd>I</kbd> (Mac) または <kbd>Ctrl</kbd>+<kbd>I</kbd> (Windows/Linux) | `_This text is italicized_`                       |                                          |                                         |                              |
| *このテキストは斜体*                                                                            |                                                   |                                          |                                         |                              |
| 取り消し線                                                                                  |                                                   |                                          |                                         |                              |
| `~~ ~~` または `~ ~`                                                                      | None                                              | `~~This was mistaken text~~`             |                                         |                              |
| ~~これは間違ったテキストでした~~                                                                     |                                                   |                                          |                                         |                              |
| 太字とネストされた斜体                                                                            |                                                   |                                          |                                         |                              |
| `** **` および `_ _`                                                                      | None                                              | `**This text is _extremely_ important**` |                                         |                              |
| **このテキストは *非常に* 重要です**                                                                 |                                                   |                                          |                                         |                              |
| すべて太字と斜体                                                                               | `*** ***`                                         | None                                     | `***All this text is important***`      |                              |
| ***このテキストはすべて重要です***                                                                   | <!-- markdownlint-disable-line emphasis-style --> |                                          |                                         |                              |
| 下付き                                                                                    | `<sub> </sub>`                                    | None                                     | `This is a <sub>subscript</sub> text`   | これは <sub>下付きテキスト</sub> です    |
| 上付き                                                                                    | `<sup> </sup>`                                    | None                                     | `This is a <sup>superscript</sup> text` | これは <sup>上付き文字</sup> のテキストです |
| 下線                                                                                     | `<ins> </ins>`                                    | None                                     | `This is an <ins>underlined</ins> text` | これは <ins>下線付き</ins> テキストです   |

## テキストの引用

<kbd>
>
</kbd>を使用してテキストを引用符で囲むことができます。

```markdown
Text that is not a quote

> Text that is a quote
```

引用されたテキストは、左に縦線を用いてインデントされ、灰色のフォントで表示されます。

![標準テキストと引用符で囲まれたテキストの違いを示す、マークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/quoted-text-rendered.png)

> \[!NOTE]
> 会話を表示するときに、テキストを強調表示してから <kbd>「R</kbd>」と入力することで、コメント内のテキストを自動的に引用符で囲むことができます。コメント全体を引用するには、\[ <svg version="1.1" width="16" height="16" viewBox="0 0 16 16" class="octicon octicon-kebab-horizontal" aria-label="The horizontal kebab icon" role="img"><path d="M8 9a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3ZM1.5 9a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3Zm13 0a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3Z"></path></svg>]、\[ **返信の見積もり**] の順にクリックします。 キーボード ショートカットの詳細については、「 [キーボード ショートカット](/ja/get-started/accessibility/keyboard-shortcuts)」を参照してください。

## コードの引用方法

1 つのバックティックを使用して、文内でコードまたはコマンドを呼び出すことができます。 バックティック内のテキストは書式設定されません。 <kbd>Command</kbd>+<kbd>E</kbd> (Mac) または <kbd>Ctrl を押すこともできます。 </kbd>+<kbd>E</kbd> (Windows/Linux) キーボード ショートカットを使用して、コード ブロックのバックティックを Markdown 行内に挿入します。

```markdown
Use `git status` to list all new or modified files that haven't yet been committed.
```

![バッククォークで囲まれた文字が固定幅の書体で表示され、淡い灰色で強調表示されていることを示すマークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/inline-code-rendered.png)

コードやテキストを独自で際立ったブロックにフォーマットするには、トリプルバックティックを使用します。

````markdown
Some basic Git commands are:
```
git status
git add
git commit
```
````

![構文が強調表示されていない単純なコード ブロックを示すマークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/code-block-rendered.png)

詳しくは、「[コードブロックの作成と強調表示](/ja/get-started/writing-on-github/working-with-advanced-formatting/creating-and-highlighting-code-blocks)」をご覧ください。

コード スニペットとテーブルを頻繁に編集する場合は、GitHub のすべてのコメント フィールドで固定幅フォントを有効にするとメリットが得られる場合があります。 詳しくは、「[GitHubでの書き込みと書式設定について](/ja/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/about-writing-and-formatting-on-github#enabling-fixed-width-fonts-in-the-editor)」をご覧ください。

## サポートされているカラー モデル

問題、プルリクエスト、ディスカッションでは、バックティックを使用して文章内の色を強調することができます。 バックティックス内でサポートされているカラー モデルでは、色の視覚化が表示されます。

```markdown
The background color is `#ffffff` for light mode and `#000000` for dark mode.
```

![バックティック内の HEX 値によって色の小さな円が作成され、ここに白と黒が表示されているGitHubマークダウンのスクリーンショット。](/assets/images/help/writing/supported-color-models-rendered.png)

現在サポートされているカラー モデルを次に示します。

| 色    | 構文                          | Example                             | アウトプット                                                                                                                                      |
| ---- | --------------------------- | ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| 16進数 | <code>\`#RRGGBB\`</code>    | <code>\`#0969DA\`</code>            | ![16 進数の値 #0969DA が青い円で表示される方法を示す、マークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/supported-color-models-hex-rendered.png)     |
| RGB  | <code>\`rgb(R,G,B)\`</code> | <code>\`rgb(9, 105, 218)\`</code>   | ![RGB 値 9、105、218 が青い円でどのように表示されるかを示す、マークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/supported-color-models-rgb-rendered.png) |
| HSL  | <code>\`hsl(H,S,L)\`</code> | <code>\`hsl(212, 92%, 45%)\`</code> | ![HSL 値 212、92%、45% が青い円で表示される方法を示す、マークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/supported-color-models-hsl-rendered.png)   |

> \[!NOTE]
>
> * サポートされているカラー モデルでは、バックティック内に先頭または末尾のスペースを含めることはできません。
> * 色の視覚化は、問題、プル要求、およびディスカッションでのみサポートされます。

## Links

インライン リンクを作成するには、リンク テキストを角かっこ `[ ]`で折り返し、url をかっこで囲んで `( )`。 キーボード ショートカット <kbd>の Command</kbd>+<kbd>K</kbd> を使用してリンクを作成することもできます。 テキストを選択したら、クリップボードから URL を貼り付けて、選択範囲からリンクを自動的に作成できます。

テキストを強調表示し、キーボード ショートカット <kbd>Command</kbd>+<kbd>V</kbd> を使用して Markdown ハイパーリンクを作成することもできます。 テキストをリンクに置き換える場合は、キーボード ショートカット <kbd>の Command</kbd>+<kbd>Shift</kbd>+<kbd>V</kbd> を使用します。

`This site was built using [GitHub Pages](https://pages.github.com/).`

![角かっこ内のテキスト "GitHub ページ" が青いハイパーリンクとしてどのように表示されるかを示す、マークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/link-rendered.png)

> \[!NOTE]
> GitHub 有効な URL がコメントに書き込まれると、リンクが自動的に作成されます。 詳しくは、「[自動リンクされた参照と URL](/ja/get-started/writing-on-github/working-with-advanced-formatting/autolinked-references-and-urls)」をご覧ください。

## セクションのリンク

見出しがある任意のセクションに直接リンクできます。 レンダリングされたファイルで自動的に生成されたアンカーを表示するには、セクション見出しにカーソルを合わせて <svg version="1.1" width="16" height="16" viewBox="0 0 16 16" class="octicon octicon-link" aria-label="the link" role="img"><path d="m7.775 3.275 1.25-1.25a3.5 3.5 0 1 1 4.95 4.95l-2.5 2.5a3.5 3.5 0 0 1-4.95 0 .751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018 1.998 1.998 0 0 0 2.83 0l2.5-2.5a2.002 2.002 0 0 0-2.83-2.83l-1.25 1.25a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042Zm-4.69 9.64a1.998 1.998 0 0 0 2.83 0l1.25-1.25a.751.751 0 0 1 1.042.018.751.751 0 0 1 .018 1.042l-1.25 1.25a3.5 3.5 0 1 1-4.95-4.95l2.5-2.5a3.5 3.5 0 0 1 4.95 0 .751.751 0 0 1-.018 1.042.751.751 0 0 1-1.042.018 1.998 1.998 0 0 0-2.83 0l-2.5 2.5a1.998 1.998 0 0 0 0 2.83Z"></path></svg> アイコンを表示し、アイコンをクリックしてブラウザーにアンカーを表示します。

![リポジトリの README のスクリーンショット。 セクション見出しの左側では、リンク アイコンが濃いオレンジ色の枠線で囲まれています。](/assets/images/help/repository/readme-links.png)

編集中のファイル内の見出しのアンカーを決定する必要がある場合は、次の基本ルールを使用できます。

* 文字は小文字に変換されます。
* スペースはハイフン (`-`) に置き換えられます。 その他の空白文字または句読点文字は削除されます。
* 先頭と末尾の空白が削除されます。
* マークアップの書式設定は削除され、内容のみが残ります (たとえば、 `_italics_` は `italics`になります)。
* 見出しに対して自動的に生成されるアンカーが同じドキュメント内の以前のアンカーと同じ場合は、ハイフンと自動インクリメント整数を追加することで一意の識別子が生成されます。

URI フラグメントの要件の詳細については、「 [RFC 3986: Uniform Resource Identifier (URI): Generic Syntax, Section 3.5](https://www.rfc-editor.org/rfc/rfc3986#section-3.5)」を参照してください。

次のコード ブロックは、レンダリングされたコンテンツの見出しからアンカーを生成するために使用される基本的なルールを示しています。

```markdown
# Example headings

## Sample Section

## This'll be a _Helpful_ Section About the Greek Letter Θ!
A heading containing characters not allowed in fragments, UTF-8 characters, two consecutive spaces between the first and second words, and formatting.

## This heading is not unique in the file

TEXT 1

## This heading is not unique in the file

TEXT 2

# Links to the example headings above

Link to the sample section: [Link Text](#sample-section).

Link to the helpful section: [Link Text](#thisll-be-a-helpful-section-about-the-greek-letter-Θ).

Link to the first non-unique section: [Link Text](#this-heading-is-not-unique-in-the-file).

Link to the second non-unique section: [Link Text](#this-heading-is-not-unique-in-the-file-1).
```

> \[!NOTE]
> 見出しを編集する場合、または "同じ" アンカーで見出しの順序を変更する場合は、アンカーが変更されるため、それらの見出しへのリンクも更新する必要があります。

## 相対リンク

表示されたファイル中で相対リンクと画像パスを定義して、読者がリポジトリ中の他のファイルにアクセスしやすくできます。

相対リンクは、現在のファイルに対する相対的なリンクです。 たとえば、リポジトリのルートに README ファイルがあり、*docs/CONTRIBUTING.md* に別のファイルがある場合、README の *CONTRIBUTING.md* への相対リンクは次のようになります。

```text
[Contribution guidelines for this project](docs/CONTRIBUTING.md)
```

GitHub は相対リンクまたは画像パスを、現在のブランチに基づいて変換するので、リンクやパスは常にうまく働きます。 リンクのパスは、現在のファイルに対する相対パスになります。 `/` で始まるリンクは、リポジトリ ルートに対する相対パスです。 `./` や `../` のような相対リンクのオペランドをすべて使用できます。

リンク テキストは 1 行に記述する必要があります。 次の例は機能しません。

```markdown
[Contribution
guidelines for this project](docs/CONTRIBUTING.md)
```

相対リンクは、リポジトリをクローンするユーザにも扱いやすいです。 絶対リンクはリポジトリのクローンではうまく働かないかもしれません。リポジトリ内の他のファイルを参照するには、相対リンクを使うことをおすすめします。

## カスタム アンカー

標準の HTML アンカー タグ (`<a name="unique-anchor-name"></a>`) を使用して、ドキュメント内の任意の場所のナビゲーション アンカー ポイントを作成できます。 あいまいな参照を回避するには、 `name` 属性値にプレフィックスを追加するなど、アンカー タグに一意の名前付けスキームを使用します。

> \[!NOTE]
> カスタム アンカーは、ドキュメントアウトライン/目次には含まれません。

アンカーに指定した `name` 属性の値を使用して、カスタム アンカーにリンクできます。 構文は、見出しに対して自動的に生成されるアンカーにリンクする場合とまったく同じです。

例えば次が挙げられます。

```markdown
# Section Heading

Some body text of this section.

<a name="my-custom-anchor-point"></a>
Some text I want to provide a direct link to, but which doesn't have its own heading.

(… more content…)

[A link to that custom anchor](#my-custom-anchor-point)
```

> \[!TIP]
> カスタム アンカーは、自動見出しリンクの自動名前付けおよび番号付け動作では考慮されません。

## 改行コード

リポジトリで問題、pull request、またはディスカッションを記述している場合、 GitHub は改行を自動的にレンダリングします。

```markdown
This example
Will span two lines
```

ただし、.md ファイルで記述している場合、上記の例は改行なしで 1 行にレンダリングされます。 .md ファイルに改行を作成するには、次のいずれかを含める必要があります。

* 最初の行の末尾に 2 つのスペースを含めます。
  <pre>
  This example&nbsp;&nbsp;
  Will span two lines
  </pre>

* 最初の行の末尾にバックスラッシュを含めます。

  ```markdown
  This example\
  Will span two lines
  ```

* 最初の行の末尾に HTML 単一改行タグを含めます。

  ```markdown
  This example<br/>
  Will span two lines
  ```

2 行の間に空白行を残した場合、.md ファイルと Markdown の両方が問題、pull request、ディスカッションで空白行で区切られた 2 行をレンダリングします。

```markdown
This example

Will have a blank line separating both lines
```

## 画像

を追加することで画像を表示できます <kbd>。</kbd> 代替テキストを `[ ]`で折り返します。 代替テキストは、画像内の情報に相当する短いテキストです。 次に、イメージのリンクをかっこ `()`で囲みます。

`![Screenshot of a comment on a GitHub issue showing an image, added in the Markdown, of an Octocat smiling and raising a tentacle.](https://myoctocat.com/assets/images/base-octocat.svg)`

![Octocat が笑顔で触手を上げている画像がマークダウンに追加されていることを示す、GitHubの問題に関するコメントのスクリーンショット。](/assets/images/help/writing/image-rendered.png)

GitHub では、問題、pull request、、ディスカッション、、コメント、 `.md` ファイルへの画像の埋め込みをサポートしています。 リポジトリからイメージを表示したり、オンライン イメージへのリンクを追加したり、イメージをアップロードしたりできます。 詳細については、「 [アセットのアップロード」を](#uploading-assets)参照してください。

> \[!NOTE]
> リポジトリ内のイメージを表示する場合は、絶対リンクではなく相対リンクを使用します。

相対リンクを使用して画像を表示する例を次に示します。

| Context                            | 相対リンク                                                                  |
| ---------------------------------- | ---------------------------------------------------------------------- |
| 同じブランチ上の `.md` ファイル内               | `/assets/images/electrocat.png`                                        |
| 別のブランチの `.md` ファイル内                | `/../main/assets/images/electrocat.png`                                |
| リポジトリの課題、プルリクエスト、コメントにおいて          | `../blob/main/assets/images/electrocat.png?raw=true`                   |
| 別のリポジトリの `.md` ファイル内               | `/../../../../github/docs/blob/main/assets/images/electrocat.png`      |
| 問題の場合は、別のリポジトリの pull request とコメント | `../../../github/docs/blob/main/assets/images/electrocat.png?raw=true` |

> \[!NOTE]
> 上記の表の最後の 2 つの相対リンクは、ビューアーが少なくともこれらのイメージを含むプライベート リポジトリへの読み取りアクセス権を持っている場合にのみ、プライベート リポジトリ内のイメージに対して機能します。

詳細については、「 [相対リンク](#relative-links)」を参照してください。

### Picture 要素

`<picture>` HTML 要素がサポートされています。

## Lists

1 行以上のテキストの前に <kbd>-</kbd>、 <kbd>\*</kbd>、または <kbd>+</kbd>を付けることで、順序付けされていないリストを作成できます。

```markdown
- George Washington
* John Adams
+ Thomas Jefferson
```

![最初の 3 人のアメリカ大統領の名前の箇条書きリストを示すマークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/unordered-list-rendered.png)

リストを並べ替える場合は、各行の前に数字を付けておきます。

```markdown
1. James Madison
2. James Monroe
3. John Quincy Adams
```

![4 番目、5 番目、6 番目のアメリカ大統領の名前の番号付きリストを示すマークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/ordered-list-rendered.png)

### 入れ子になったリスト

1 つ以上のリスト 項目を別の項目の下にインデントすることで、入れ子になったリストを作成できます。

GitHubの Web エディターを使用して入れ子になったリストを作成したり、[Visual Studio Code](https://code.visualstudio.com/)などの単一スペース フォントを使用するテキスト エディターを作成したりするには、リストを視覚的に配置できます。 リスト マーカー文字 (<kbd>-</kbd> または <kbd>\*</kbd>) が、その上の項目のテキストの最初の文字のすぐ下に配置されるまで、入れ子になったリスト アイテムの前にスペース文字を入力します。

```markdown
1. First list item
   - First nested list item
     - Second nested list item
```

> \[!NOTE]
> Web ベースのエディターでは、最初に目的の行を強調表示してから <kbd>Tab</kbd> または <kbd>Shift</kbd>+<kbd>Tab</kbd> を使用して、1 行以上のテキストをインデントまたはデデントできます。

![スクリーンショットの Markdown Visual Studio Code では、入れ子になった番号付き行と箇条書きのインデントが示されています。](/assets/images/help/writing/nested-list-alignment.png)

![2 つの異なるレベルの入れ子になった項目の後に入れ子になった行頭文字が続くマークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/nested-list-example-1.png)

モノスペース フォントを使用しない GitHubのコメント エディターで入れ子になったリストを作成するには、入れ子になったリストのすぐ上にあるリスト アイテムを確認し、アイテムのコンテンツの前に表示される文字数をカウントします。 次に、入れ子になったリスト項目の前にその数の空白文字を入力します。

この例では、入れ子になったリスト アイテムを 5 文字以上インデントすることで、リストアイテム`100. First list item`の下に入れ子になったリストアイテムを追加できます。これは、`100. `前に 5 文字 (`First list item`) があるためです。

```markdown
100. First list item
     - First nested list item
```

![マークダウンGitHubレンダリングされたスクリーンショット。番号 100 で始まり、1 レベルで入れ子になった箇条書きの項目が続く番号付きの項目が示されています。](/assets/images/help/writing/nested-list-example-3.png)

同じメソッドを使用して、入れ子になったリストの複数のレベルを作成できます。 たとえば、入れ子になったリスト アイテムの先頭には、入れ子になったリスト コンテンツ`␣␣␣␣␣-␣`の前に 7 文字 (`First nested list item`) があるため、2 番目の入れ子になったリストアイテムを少なくとも 2 文字インデントする必要があります (最小 9 文字)。

```markdown
100. First list item
     - First nested list item
       - Second nested list item
```

![マークダウンGitHubレンダリングされたスクリーンショット。番号 100 で始まり、2 つの異なるレベルの入れ子の行頭文字が続く番号付きの項目が示されています。](/assets/images/help/writing/nested-list-example-2.png)

その他の例については、[GitHub Flavored Markdown Spec](https://github.github.com/gfm/#example-265) を参照してください。

## タスク リスト

タスク リストを作成するには、リスト アイテムの前に空白、ハイフン、`[ ]` を付けます。 完了したタスクをマークするには、`[x]` を使います。

```markdown
- [x] #739
- [ ] https://github.com/octo-org/octo-repo/issues/740
- [ ] Add delight to the experience when all tasks are complete :tada:
```

![マークダウンのレンダリング バージョンを示すスクリーンショット。 issue の参照が issue のタイトルとしてレンダリングされています。](/assets/images/help/writing/task-list-rendered-simple.png)

タスク リスト アイテムの説明がかっこで始まる場合は、 <kbd>\\</kbd>でエスケープする必要があります。

`- [ ] \(Optional) Open a followup issue`

詳しくは、「[タスクリストについて](/ja/get-started/writing-on-github/working-with-advanced-formatting/about-tasklists)」をご覧ください。

## 人々やチームのメンション

ユーザー名または[チーム](/ja/organizations/organizing-members-into-teams)名GitHub入力すると、@でユーザーまたはチームにメンションできます。 これにより通知がトリガーされ、会話に注意が向けられます。 ユーザー名またはチーム名をメンションするようにコメントを編集すると、ユーザーにも通知が届きます。 通知の詳細については、「[通知について](/ja/subscriptions-and-notifications/concepts/about-notifications)」を参照してください。

> \[!NOTE]
> ユーザーには、そのユーザーがリポジトリへの読み取りアクセス権を持ち、リポジトリが組織によって所有されている場合は、そのユーザーが組織のメンバーである場合にのみ、メンションに関する通知が送信されます。

`@github/support What do you think about these updates?`

![チームメンション "@github/support" が太字でクリック可能なテキストとしてレンダリングされる方法を示す、マークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/mention-rendered.png)

親チームに言及すると、その子チームのメンバーも通知を受け取り、複数のユーザー グループとのコミュニケーションが簡素化されます。 詳しくは、「[Organization のチームについて](/ja/organizations/organizing-members-into-teams/about-teams)」をご覧ください。

<kbd>
@
</kbd>記号を入力すると、プロジェクトのユーザーまたはチームの一覧が表示されます。 リストは入力時にフィルター処理されるため、探しているユーザーまたはチームの名前を見つけたら、方向キーを使用して選択し、Tab キーまたは Enter キーを押して名前を完成させることができます。 チームの場合は、 @organization/team-name を入力すると、そのチームのすべてのメンバーが会話にサブスクライブされます。

オートコンプリートの結果は、リポジトリコラボレーターとスレッド上の他のすべての参加者に制限されます。

## 問題と pull request の参照

「 <kbd>#</kbd>」と入力すると、リポジトリ内で推奨される問題とプル要求の一覧を表示できます。 問題または pull request 番号またはタイトルを入力して一覧をフィルター処理し、Tab キーまたは Enter キーを押して強調表示された結果を完了します。

詳しくは、「[自動リンクされた参照と URL](/ja/get-started/writing-on-github/working-with-advanced-formatting/autolinked-references-and-urls)」をご覧ください。

## 外部リソースの参照

カスタムの自動リンク参照がリポジトリに設定されているなら、JIRAのIssueやZendeskのチケットのような外部リソースへの参照は、短縮リンクに変換されます。 リポジトリで利用できる自動リンクを知るには、リポジトリの管理権限を持つ人に連絡してください。 詳しくは、「[外部リソースを参照する自動リンクの構成](/ja/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings/configuring-autolinks-to-reference-external-resources)」をご覧ください。

## アセットのアップロード

画像などのアセットは、ドラッグ アンド ドロップ、ファイル ブラウザーからの選択、貼り付けによってアップロードできます。 リポジトリ内の問題、pull request、コメント、 `.md` ファイルにアセットをアップロードできます。

## 絵文字の使用

`:EMOJICODE:`入力し、コロンの後に絵文字の名前を付けることで、絵文字を文書に追加できます。

`@octocat :+1: This PR looks great - it's ready to merge! :shipit:`

![+1 と shipit の絵文字コードが絵文字として視覚的にレンダリングされる方法を示す、レンダリングされたGitHub Markdown のスクリーンショット。](/assets/images/help/writing/emoji-rendered.png)

入力 <kbd>:</kbd> 提案された絵文字のリストが表示されます。 入力するとリストがフィルター処理されるため、探している絵文字が見つかると、 **Tab** キーまたは **Enter** キーを押して強調表示された結果を完了します。

利用可能な絵文字とコードの完全な一覧については、 [絵文字チートシートを](https://github.com/ikatyang/emoji-cheat-sheet/blob/github-actions-auto-update/README.md)参照してください。

## 段落

テキスト行間に空白行を残すことで、新しい段落を作成できます。

## 脚注

次の角かっこ構文を使用して、コンテンツに脚注を追加できます。

```text
Here is a simple footnote[^1].

A footnote can also have multiple lines[^2].

[^1]: My reference.
[^2]: To add line breaks within a footnote, add 2 spaces to the end of a line.  
This is a second line.
```

脚注は次のようにレンダリングされます。

![マークダウンがレンダリングされたスクリーンショットで、脚注を示すための上付き文字と、ノート内で使えるオプションの改行を表示しています。](/assets/images/help/writing/footnote-rendered.png)

> \[!NOTE]
> マークダウン内の脚注の位置は、脚注がレンダリングされる場所には影響しません。 脚注への参照の直後に脚注を書くことができますが、脚注は Markdown の下部に引き続きレンダリングされます。 脚注は Wiki ではサポートされていません。

## Alerts

**アラート** ( **吹き出し** や **警告**とも呼ばれます) は、重要な情報を強調するために使用できるブロッククォート構文に基づく Markdown 拡張機能です。
GitHubでは、コンテンツの重要性を示す独特の色とアイコンが表示されます。

アラートは、ユーザーの成功に不可欠な場合にのみ使用し、読者の過負荷を防ぐために記事ごとに 1 つまたは 2 つに制限します。 さらに、アラートを連続して配置しないようにする必要があります。 アラートを他の要素内に入れ子にすることはできません。

アラートを追加するには、アラートの種類を指定する特殊なブロッククォート行を使用し、その後に標準ブロッククォート内のアラート情報を指定します。 次の 5 種類のアラートを使用できます。

```markdown
> [!NOTE]
> Useful information that users should know, even when skimming content.

> [!TIP]
> Helpful advice for doing things better or more easily.

> [!IMPORTANT]
> Key information users need to know to achieve their goal.

> [!WARNING]
> Urgent info that needs immediate user attention to avoid problems.

> [!CAUTION]
> Advises about risks or negative outcomes of certain actions.
```

表示されるアラートを次に示します。

![さまざまな色付きのテキストとアイコンを使用してメモ、ヒント、重要、警告、および注意がどのようにレンダリングされるかを示すマークダウン アラートのレンダリングのスクリーンショット。](/assets/images/help/writing/alerts-rendered.png)

## コメントを含むコンテンツを非表示にする

コンテンツを HTML コメントに配置することで、レンダリングされた Markdown からコンテンツを非表示にするように GitHub に指示できます。

```text
<!-- This content will not appear in the rendered Markdown -->
```

## マークダウンの書式設定を無視する

Markdown 文字の前にGitHubを使用して、マークダウンの書式設定を無視 (またはエスケープ) するように\に指示できます。

`Let's rename \*our-new-project\* to \*our-old-project\*.`

![バックスラッシュによってアスタリスクが斜体に変換されないようにする方法を示す、マークダウンGitHubレンダリングされたスクリーンショット。](/assets/images/help/writing/escaped-character-rendered.png)

円記号の詳細については、「Daring Fireball の [マークダウン構文](https://daringfireball.net/projects/markdown/syntax#backslash)」を参照してください。

> \[!NOTE]
> マークダウンの書式設定は、問題またはプル要求のタイトルでは無視されません。

## Markdown レンダリングの無効化

Markdown ファイルを表示するときに、ファイルの上部にある **\[Code]** をクリックすると、Markdown レンダリングが無効になり、代わりにファイルのソースが表示されます。

![ファイルを操作するためのオプションが示されているリポジトリ内のマークダウン ファイルのスクリーンショット。 \[コード\] というラベルが付いたボタンが濃いオレンジ色の枠線で囲まれています。](/assets/images/help/writing/display-markdown-as-source-global-nav-update.png)

Markdown レンダリングを無効にすると、ライン リンクなどのソース ビュー機能を使用できます。これは、レンダリングされた Markdown ファイルを表示する場合には使用できません。

## 詳細については、次を参照してください。

* [
  GitHub フレーバーマークダウン仕様](https://github.github.com/gfm/)
* [GitHubでの書き込みと書式設定について](/ja/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/about-writing-and-formatting-on-github)
* [高度なフォーマットを使用して作業する](/ja/get-started/writing-on-github/working-with-advanced-formatting)
* [GitHubでの記述に関するクイック スタート](/ja/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/quickstart-for-writing-on-github)
