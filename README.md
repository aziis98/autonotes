# AutoNotes Project

A Go-based tool for **agent-assisted transcription** of handwritten mathematical notes into a structured digital format, generating a professional, interactive HTML visualization.

## Agent-Assisted Transcription

This project is explicitly designed to be operated by AI agents. Agents are responsible for viewing source images, extracting mathematical content, and mapping specific regions (using a 1000x1000 coordinate system) into `.note` files.

For detailed instructions on the transcription protocol, agents **must** refer to [AGENTS.md](./AGENTS.md).

## Folder Structure

- `src/`: Contains the source materials.
  - `[collection-name]/`: E.g., `ist-geom/`.
    - `images/`: The original handwritten photos (JPEG).
    - `[filename].note`: The structured transcription files.
- `out/`: The generated standalone website.
- `main.go`, `build.go`, `status.go`, `sync.go`, `parser.go`: The core CLI and logic.
- `AGENTS.md`: Detailed workflow and instructions for transcription agents.

## Usage

Run the sync command to build and check for new work:

```bash
go run -v . sync
```

## Syntax Overview (.note files)

Note files use an XML-like syntax to map transcriptions to images using a **1000x1000 coordinate system**.

### Base Elements

- `<box image="name.jpg" top="Y" right="X" bottom="Y" left="X">`: Maps transcribed text to a region on an image. Triggers a red highlight and lens view on hover.
- `<math>` and `<math display="true">`: KaTeX-powered mathematical expressions.
- `<reword>`: Formal textbook-style rewrite of the transcription.
- `<image src="name.jpg" top="..." right="..." bottom="..." left="..." />`: Inline cropped diagram. Also supports hover highlighting and lens view.

### Structural Blocks

- `<definition>`, `<theorem>`, `<lemma>`, `<oss>`, `<dim>`, `<richiami>`: Semantic containers for mathematical content.
- `<itemize>`, `<enumerate>`, `<item>`: Support for bulleted and numbered lists.

## TODO

- [ ] Convert the syntax to real XML, not XML-like

- [ ] Refactor the `images/` folder concept to a `sources/` folder that can contain both images, zip of images, pdfs, etc. There is a command called `go run . extract` that unpacks archives and pdfs to a folder with a similar name that only contains image files. When generating the website, everything is flattened, given uuids, and pdfs are converted to images (with mutools).

- [ ] Add the concept of symbols, also embed in the math content, e.g. `<math><ref id="symbol-name">\alpha</ref></math>`. When converting to HTML, the refs get replaced with `\href` or more complicated `\htmlData` tags (katex feature).
  - **Global Cross-Referencing**: The `<ref id="...">...</ref>` tag is a universal linking mechanism that can wrap math symbols or plain text, creating an interconnected web of concepts.
  - **Symbol Definitions**:
    - Use `<symbol id="..." name="..." description="..." latex="..." />` to define global identifiers.
    - Definitions can be centralized in a `symbols.note` or declared at the start of a chapter.
  - **Enhanced Examples**:
    - **Constants**: `<math><ref id="const-e">e</ref>^{i\pi} + 1 = 0</math>` links to the definition of the Euler's number.
    - **Formal Terms**: `Il <ref id="def-omeomorfismo">omeomorfismo</ref> tra i due spazi...` links to the topological definition.
    - **Theorems**: `Per il <ref id="theorem-cauchy">Teorema di Cauchy</ref>...` links to the statement and its proof.
  - **Interactive HTML Delivery**:
    - **Contextual Tooltips**: Hovering over a reference displays a sleek floating card with the target's name, description, and LaTeX preview.
    - **Back-references**: Each definition automatically tracks and lists all the places where it is referenced across the collection.
    - **Navigation**: Deep links that scroll to the exact block (theorem, definition) or highlight the specific occurrence.
  - **Agent Tooling**:
    - **Symbol Inspection**: A subcommand `go run . symbols` will provide a CLI interface for agents to list all defined symbols, verify their metadata, and trace their cross-reference graph (where they are used and what they link to).

- [ ] Add a search bar using FuseJS

---

For more detailed instructions on transcription, see [AGENTS.md](./AGENTS.md).
