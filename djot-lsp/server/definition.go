package server

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func handleDefinition(ctx *glsp.Context, params *protocol.DefinitionParams) (any, error) {
	doc := getDocument(params.TextDocument.URI)
	if doc == nil {
		return nil, nil
	}
	loc := computeDefinition(doc, params.Position)
	if loc == nil {
		return nil, nil
	}
	return loc, nil
}

// computeDefinition returns the Location of the definition for whatever is
// under pos, or nil if no definition can be found.
func computeDefinition(doc *Document, pos protocol.Position) *protocol.Location {
	offset := doc.PositionToOffset(pos)
	if offset > len(doc.Source) {
		offset = len(doc.Source)
	}

	if label := findFootnoteRefAt(doc, offset); label != "" {
		if def, ok := doc.FootnoteDefs[label]; ok {
			return &protocol.Location{
				URI:   doc.URI,
				Range: def.Range,
			}
		}
		return nil
	}

	if label := findLinkRefAt(doc, offset); label != "" {
		if def, ok := doc.ReferenceDefs[label]; ok {
			return &protocol.Location{
				URI:   doc.URI,
				Range: def.Range,
			}
		}
		return nil
	}

	return nil
}
