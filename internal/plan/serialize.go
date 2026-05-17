package plan

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const (
	header        = "## commit-composer plan v1"
	rewordSep     = " :: "
	rewordB64Tag  = "b64::"
	keyBase       = "base:"
	keyRange      = "range:"
	keyOpsBegin   = "ops:"
	opLinePrefix  = "- "
	maxLineLength = 1 << 20 // 1 MiB safety bound
)

// Marshal serializes a Plan to the line-based v1 contract.
//
// Format:
//
//	## commit-composer plan v1
//	base: <full-sha>
//	range: <base>..<head>
//	ops:
//	- pick   <sha>
//	- reword <sha> :: new message
//	- reword <sha> b64::<base64-of-multiline>
//	- squash <sha>
//	- drop   <sha>
//
// Reword messages containing newlines are base64-encoded to keep the format
// line-based.
func Marshal(p Plan) string {
	var b strings.Builder
	b.WriteString(header)
	b.WriteByte('\n')
	fmt.Fprintf(&b, "%s %s\n", keyBase, p.Base)
	if p.Range != "" {
		fmt.Fprintf(&b, "%s %s\n", keyRange, p.Range)
	}
	b.WriteString(keyOpsBegin)
	b.WriteByte('\n')
	for _, op := range p.Ops {
		writeOp(&b, op)
	}
	return b.String()
}

func writeOp(b *strings.Builder, op Op) {
	b.WriteString(opLinePrefix)
	b.WriteString(op.Action.String())
	b.WriteByte(' ')
	b.WriteString(op.SHA)
	if op.Action == Reword && op.NewMessage != "" {
		b.WriteString(rewordSep)
		if strings.ContainsRune(op.NewMessage, '\n') {
			b.WriteString(rewordB64Tag)
			b.WriteString(base64.StdEncoding.EncodeToString([]byte(op.NewMessage)))
		} else {
			b.WriteString(op.NewMessage)
		}
	}
	b.WriteByte('\n')
}

// Unmarshal parses the v1 contract back into a Plan. It is tolerant of
// trailing whitespace and ignores unknown lines outside the ops block.
func Unmarshal(r io.Reader) (Plan, error) {
	var p Plan
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), maxLineLength)
	sawHeader := false
	inOps := false
	idx := 0
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t\r")
		if line == "" {
			continue
		}
		if !sawHeader {
			if line == header {
				sawHeader = true
				continue
			}
			return Plan{}, fmt.Errorf("missing plan header %q", header)
		}
		if !inOps {
			switch {
			case strings.HasPrefix(line, keyBase):
				p.Base = strings.TrimSpace(strings.TrimPrefix(line, keyBase))
			case strings.HasPrefix(line, keyRange):
				p.Range = strings.TrimSpace(strings.TrimPrefix(line, keyRange))
			case line == keyOpsBegin:
				inOps = true
			default:
				return Plan{}, fmt.Errorf("unexpected line before ops: %q", line)
			}
			continue
		}
		op, err := parseOpLine(line)
		if err != nil {
			return Plan{}, err
		}
		op.OrigIndex = idx
		idx++
		p.Ops = append(p.Ops, op)
	}
	if err := scanner.Err(); err != nil {
		return Plan{}, fmt.Errorf("scan plan: %w", err)
	}
	if !sawHeader {
		return Plan{}, fmt.Errorf("empty plan")
	}
	if p.Base == "" {
		return Plan{}, fmt.Errorf("plan missing base")
	}
	return p, nil
}

func parseOpLine(line string) (Op, error) {
	if !strings.HasPrefix(line, opLinePrefix) {
		return Op{}, fmt.Errorf("op line must start with %q: %q", opLinePrefix, line)
	}
	rest := strings.TrimPrefix(line, opLinePrefix)
	// Split action from the rest.
	actionTok, after, ok := cutFirstField(rest)
	if !ok {
		return Op{}, fmt.Errorf("op line missing action: %q", line)
	}
	action, err := ParseAction(actionTok)
	if err != nil {
		return Op{}, fmt.Errorf("op line %q: %w", line, err)
	}
	shaTok, tail, _ := cutFirstField(after)
	if shaTok == "" {
		return Op{}, fmt.Errorf("op line missing sha: %q", line)
	}
	op := Op{SHA: shaTok, Action: action}
	tail = strings.TrimLeft(tail, " \t")
	if tail == "" {
		return op, nil
	}
	if !strings.HasPrefix(tail, strings.TrimSpace(rewordSep)) {
		return Op{}, fmt.Errorf("op line trailing data not separated by %q: %q", rewordSep, line)
	}
	tailContent := strings.TrimPrefix(tail, strings.TrimSpace(rewordSep))
	tailContent = strings.TrimLeft(tailContent, " \t")
	switch op.Action {
	case Reword:
		if strings.HasPrefix(tailContent, rewordB64Tag) {
			decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(tailContent, rewordB64Tag))
			if err != nil {
				return Op{}, fmt.Errorf("op line %q: bad base64: %w", line, err)
			}
			op.NewMessage = string(decoded)
		} else {
			op.NewMessage = tailContent
		}
	case ClaudeRecompose:
		// Backward-compat: accept and ignore a trailing granularity token
		// from old "claude-split :: by-file" plans.
		_ = tailContent
	default:
		return Op{}, fmt.Errorf("op line %q: action %s does not take trailing content", line, op.Action)
	}
	return op, nil
}

// cutFirstField splits s into the first whitespace-delimited token and the
// remainder. It treats runs of spaces/tabs as a single separator.
func cutFirstField(s string) (head, tail string, ok bool) {
	s = strings.TrimLeft(s, " \t")
	if s == "" {
		return "", "", false
	}
	i := strings.IndexAny(s, " \t")
	if i < 0 {
		return s, "", true
	}
	return s[:i], s[i+1:], true
}
