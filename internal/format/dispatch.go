package format

import (
	"github.com/pawnkit/pawn-parser"
	"github.com/pawnkit/pawnfmt/internal/config"
	"github.com/pawnkit/pawnfmt/internal/doc"
)

func (s *state) formatNode(n *parser.Node) doc.Doc {
	if n == nil {
		return doc.Text("")
	}
	if n.Kind == parser.KindSourceFile {
		return s.dispatch(n)
	}
	if n.Kind == parser.KindConditionalRegion {
		return s.dispatch(n)
	}
	var core doc.Doc
	if n.Kind == parser.KindRaw || n.HasError || s.isDisabled(n) {
		core = s.withTrivia(n, s.raw(n))
	} else {
		core = s.withTrivia(n, s.dispatch(n))
	}
	if n.Kind.IsDirective() && s.config.DirectiveIndent == config.DirectiveIndentNone {
		return doc.ResetIndent(core)
	}
	return core
}

func (s *state) itemCore(n *parser.Node) doc.Doc {
	if n == nil {
		return doc.Text("")
	}
	lead := s.leadingDocs(n.LeadingTrivia())
	var core doc.Doc
	if n.Kind == parser.KindRaw || n.HasError {
		core = s.raw(n)
	} else {
		core = s.dispatch(n)
	}
	if len(lead) == 0 {
		return core
	}
	parts := append(append([]doc.Doc{}, lead...), core)
	return doc.Concat(parts...)
}

func (s *state) formatListItem(n *parser.Node, addComma bool) doc.Doc {
	parts := []doc.Doc{s.itemCore(n)}
	if addComma {
		parts = append(parts, doc.Text(","))
	}
	if trail := s.trailingDoc(n.TrailingTrivia()); trail != nil {
		parts = append(parts, trail)
	}
	return doc.Concat(parts...)
}

func (s *state) formatLastListItem(n *parser.Node) doc.Doc {
	if s.config.TrailingComma != config.TrailingCommaMultiline {
		return s.formatListItem(n, false)
	}
	parts := []doc.Doc{s.itemCore(n), doc.IfBreak(doc.Text(","), doc.Text(""))}
	if trail := s.trailingDoc(n.TrailingTrivia()); trail != nil {
		parts = append(parts, trail)
	}
	return doc.Concat(parts...)
}

func (s *state) withTrivia(n *parser.Node, core doc.Doc) doc.Doc {
	lead := s.leadingDocs(n.LeadingTrivia())
	trail := s.trailingDoc(n.TrailingTrivia())
	if len(lead) == 0 && trail == nil {
		return core
	}
	parts := make([]doc.Doc, 0, len(lead)+2)
	parts = append(parts, lead...)
	parts = append(parts, core)
	if trail != nil {
		parts = append(parts, trail)
	}
	return doc.Concat(parts...)
}

func (s *state) raw(n *parser.Node) doc.Doc {
	return doc.RawTextBlock(n.Text(s.source))
}

func (s *state) isDisabled(n *parser.Node) bool {
	if s.config.FormatDisabledRegions {
		return false
	}
	return s.trivia.OverlapsDisabled(uint32(n.Start), uint32(n.End)) //nolint:gosec // Pawn source files stay well under 4GB; offsets cannot overflow uint32
}

func (s *state) dispatch(n *parser.Node) doc.Doc {
	if formatted, ok := s.dispatchTopLevel(n); ok {
		return formatted
	}
	if formatted, ok := s.dispatchDeclaration(n); ok {
		return formatted
	}
	if formatted, ok := s.dispatchStatement(n); ok {
		return formatted
	}
	if formatted, ok := s.dispatchExpression(n); ok {
		return formatted
	}
	return s.raw(n)
}

func (s *state) dispatchTopLevel(n *parser.Node) (doc.Doc, bool) {
	switch n.Kind {
	case parser.KindSourceFile:
		return s.formatSourceFile(n), true
	case parser.KindConditionalRegion:
		return s.formatConditionalRegion(n), true
	case parser.KindSharedConditional, parser.KindSharedConditionalPrefix, parser.KindConditionalSplice:
		return s.formatSharedConditional(n), true
	case parser.KindConditionalFunction:
		return s.formatConditionalFunctionDefinition(n), true
	case parser.KindDirectiveInclude, parser.KindDirectiveTryInclude:
		return s.formatIncludeDirective(n), true
	case parser.KindDirectiveDefine:
		return s.formatDefineDirective(n), true
	case parser.KindDirectiveIf, parser.KindDirectiveElseif, parser.KindDirectiveAssert:
		return s.formatConditionDirective(n), true
	case parser.KindDirectiveUndef, parser.KindDirectivePragma, parser.KindDirectiveError,
		parser.KindDirectiveWarning, parser.KindDirectiveEmit,
		parser.KindDirectiveLine, parser.KindDirectiveFile, parser.KindDirectiveEndinput,
		parser.KindDirectiveRaw, parser.KindDirectiveElse, parser.KindDirectiveEndif:
		return s.formatRawDirectiveLine(n), true
	default:
		return nil, false
	}
}

func (s *state) dispatchDeclaration(n *parser.Node) (doc.Doc, bool) {
	switch n.Kind {
	case parser.KindFunctionDefinition, parser.KindFunctionDeclaration:
		return s.formatFunction(n), true
	case parser.KindVariableDeclaration:
		return s.formatVariableDeclaration(n), true
	case parser.KindVariableDeclarator:
		return s.formatVariableDeclarator(n), true
	case parser.KindEnumDeclaration:
		return s.formatEnumDeclaration(n), true
	case parser.KindEnumEntry:
		return s.formatEnumEntry(n), true
	case parser.KindParameter:
		return s.formatParameter(n), true
	default:
		return nil, false
	}
}

func (s *state) dispatchStatement(n *parser.Node) (doc.Doc, bool) {
	switch n.Kind {
	case parser.KindBlock:
		return s.formatBlock(n), true
	case parser.KindIfStatement:
		return s.formatIfStatement(n), true
	case parser.KindWhileStatement:
		return s.formatWhileStatement(n), true
	case parser.KindDoWhileStatement:
		return s.formatDoWhileStatement(n), true
	case parser.KindForStatement:
		return s.formatForStatement(n), true
	case parser.KindSwitchStatement:
		return s.formatSwitchStatement(n), true
	case parser.KindCaseClause, parser.KindDefaultClause:
		return s.formatCaseClause(n), true
	case parser.KindCaseValueList:
		return s.formatCaseValueList(n), true
	case parser.KindCaseRange:
		return s.formatCaseRange(n), true
	case parser.KindGotoStatement:
		return s.formatSimpleTrailingSemi(n, "goto ", "label"), true
	case parser.KindLabelStatement:
		return s.formatLabelStatement(n), true
	case parser.KindReturnStatement:
		return s.formatReturnStatement(n), true
	case parser.KindBreakStatement:
		return doc.Concat(doc.Text("break"), semiDoc(n)), true
	case parser.KindContinueStatement:
		return doc.Concat(doc.Text("continue"), semiDoc(n)), true
	case parser.KindStateStatement:
		return s.formatSimpleTrailingSemi(n, "state ", "state"), true
	case parser.KindExpressionStatement:
		return s.formatExpressionStatement(n), true
	case parser.KindEmptyStatement:
		return doc.Text(";"), true
	case parser.KindMacroInvocationBlock:
		return s.formatMacroInvocationBlock(n), true
	default:
		return nil, false
	}
}

func (s *state) dispatchExpression(n *parser.Node) (doc.Doc, bool) {
	switch n.Kind {
	case parser.KindIdentifier, parser.KindLiteral, parser.KindArgumentName, parser.KindIteratorArgument:
		return doc.Text(n.Text(s.source)), true
	case parser.KindCallExpression:
		return s.formatCallExpression(n), true
	case parser.KindSubscriptExpression:
		return s.formatSubscriptExpression(n), true
	case parser.KindTernaryExpression:
		return s.formatTernaryExpression(n), true
	case parser.KindBinaryExpression:
		return s.formatBinaryExpression(n), true
	case parser.KindUnaryExpression:
		return s.formatUnaryExpression(n), true
	case parser.KindUpdateExpression:
		return s.formatUpdateExpression(n), true
	case parser.KindAssignmentExpression:
		return s.formatAssignmentExpression(n), true
	case parser.KindSizeofExpression:
		return s.formatSizeofLikeExpression(n, "sizeof"), true
	case parser.KindTagofExpression:
		return s.formatSizeofLikeExpression(n, "tagof"), true
	case parser.KindDefinedExpression:
		return s.formatDefinedExpression(n), true
	case parser.KindTaggedExpression:
		return s.formatTaggedExpression(n), true
	case parser.KindParenthesizedExpression:
		return s.formatParenthesizedExpression(n), true
	case parser.KindArrayLiteral:
		return s.formatArrayLiteral(n), true
	case parser.KindExpressionList:
		return s.formatExpressionList(n), true
	case parser.KindStringizeExpression:
		return doc.Text(n.Text(s.source)), true
	case parser.KindStringConcat:
		return s.formatStringConcat(n), true
	default:
		return nil, false
	}
}
