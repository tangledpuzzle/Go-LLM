package textsplitter

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

// SplitDocuments splits documents using a textsplitter.
func SplitDocuments(textSplitter TextSplitter, documents []schema.Document) ([]schema.Document, error) {
	texts := make([]string, 0)
	metadatas := make([]map[string]any, 0)
	for _, document := range documents {
		texts = append(texts, document.PageContent)
		metadatas = append(metadatas, document.Metadata)
	}

	return CreateDocuments(textSplitter, texts, metadatas)
}

// CreateDocuments creates documents from texts and metadatas with a text splitter. If
// the length of the metadatas is zero or metadatas is nil, the result documents will
// contain no metadata. Otherwise the numbers of texts and metadatas must match.
func CreateDocuments(textSplitter TextSplitter, texts []string, metadatas []map[string]any) ([]schema.Document, error) {
	if len(metadatas) == 0 {
		metadatas = make([]map[string]any, len(texts))
	}

	if len(texts) != len(metadatas) {
		return nil, errors.New("number of texts and metadatas does not match")
	}

	documents := make([]schema.Document, 0)

	for i := 0; i < len(texts); i++ {
		chunks, err := textSplitter.SplitText(texts[i])
		if err != nil {
			return nil, err
		}

		for _, chunk := range chunks {
			// Copy the document metadata
			curMetadata := make(map[string]any, len(metadatas[i]))
			for key, value := range metadatas[i] {
				curMetadata[key] = value
			}

			documents = append(documents, schema.Document{
				PageContent: chunk,
				Metadata:    curMetadata,
			})
		}
	}

	return documents, nil
}

// joinDocs comines two documents with the separator used to split them.
func joinDocs(docs []string, separator string) string {
	return strings.TrimSpace(strings.Join(docs, separator))
}

// mergeSplits merges smaller splits into splits that are closer to the chunkSize.
func mergeSplits(splits []string, separator string, chunkSize int, chunkOverlap int) []string {
	fmt.Print("merging: [")
	for _, split := range splits {
		fmt.Print(split + ",")
	}
	fmt.Println("]")
	fmt.Println("seperator len ", len(separator))
	docs := make([]string, 0)
	currentDoc := make([]string, 0)
	total := 0

	for _, split := range splits {
		fmt.Println("current doc start", currentDoc)
		sepLen := len(separator)
		if len(currentDoc) == 0 {
			sepLen = 0
		}

		fmt.Println("total+len(split)+sepLen > chunkSize:", total+len(split)+sepLen > chunkSize, total, len(split), sepLen, chunkSize)
		if total+len(split)+sepLen > chunkSize {
			if total > chunkSize {
				log.Printf(
					"[WARN] created a chunk with size of %v, which is longer then the specified %v\n",
					total,
					chunkSize,
				)
			}

			if len(currentDoc) > 0 {
				doc := joinDocs(currentDoc, separator)
				if doc != "" {
					docs = append(docs, doc)
				}

				for shouldPop(chunkOverlap, chunkSize, total, len(split), len(separator), len(currentDoc)) {
					fmt.Println("cur doc before pop", currentDoc)
					total -= len(currentDoc[0]) + sepLen
					currentDoc = currentDoc[1:]
					fmt.Println("cur doc after pop", currentDoc)
				}
			}
		}

		currentDoc = append(currentDoc, split)
		sepLen = len(separator)
		if len(currentDoc) < 2 {
			sepLen = 0
		}
		fmt.Println("adding to total ", len(split)+sepLen, "len", len(split), "sep", sepLen)
		total += len(split) + sepLen
	}

	doc := joinDocs(currentDoc, separator)
	if doc != "" {
		docs = append(docs, doc)
	}

	fmt.Print("result: [")
	for _, doc := range docs {
		fmt.Print(doc + ",")
	}
	fmt.Println("]")
	return docs
}

// Keep poping if:
//   - the chunk is larger then the chunk overlap
//   - or if there are any chunks and the length is long
func shouldPop(chunkOverlap, chunkSize, total, splitLen, separatorLen, currentDocLen int) bool {
	if currentDocLen < 2 {
		separatorLen = 0
	}

	return currentDocLen > 0 && (total > chunkOverlap || (total+splitLen+separatorLen > chunkSize && total > 0))
}
