package chezmoi

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/coreos/go-semver/semver"
	vfs "github.com/twpayne/go-vfs"
	"go.uber.org/multierr"
)

// A SourceState is a source state.
type SourceState struct {
	entries         map[string]SourceStateEntry
	system          System
	sourcePath      string
	umask           os.FileMode
	encryptionTool  EncryptionTool
	ignore          *PatternSet
	minVersion      *semver.Version
	remove          *PatternSet
	templateData    map[string]interface{}
	templateFuncs   template.FuncMap
	templateOptions []string
	templates       map[string]*template.Template
}

// A SourceStateOption sets an option on a source state.
type SourceStateOption func(*SourceState)

// WithEncryptionTool set the encryption tool.
func WithEncryptionTool(encryptionTool EncryptionTool) SourceStateOption {
	return func(s *SourceState) {
		s.encryptionTool = encryptionTool
	}
}

// WithSourcePath sets the source path.
func WithSourcePath(sourcePath string) SourceStateOption {
	return func(s *SourceState) {
		s.sourcePath = sourcePath
	}
}

// WithSystem sets the system.
func WithSystem(system System) SourceStateOption {
	return func(s *SourceState) {
		s.system = system
	}
}

// WithTemplateData sets the template data.
func WithTemplateData(templateData map[string]interface{}) SourceStateOption {
	return func(s *SourceState) {
		s.templateData = templateData
	}
}

// WithTemplateFuncs sets the template functions.
func WithTemplateFuncs(templateFuncs template.FuncMap) SourceStateOption {
	return func(s *SourceState) {
		s.templateFuncs = templateFuncs
	}
}

// WithTemplateOptions sets the template options.
func WithTemplateOptions(templateOptions []string) SourceStateOption {
	return func(s *SourceState) {
		s.templateOptions = templateOptions
	}
}

// WithUmask sets the umask.
func WithUmask(umask os.FileMode) SourceStateOption {
	return func(s *SourceState) {
		s.umask = umask
	}
}

// NewSourceState creates a new source state with the given options.
func NewSourceState(options ...SourceStateOption) *SourceState {
	s := &SourceState{
		entries:         make(map[string]SourceStateEntry),
		umask:           DefaultUmask,
		encryptionTool:  &nullEncryptionTool{},
		ignore:          NewPatternSet(),
		remove:          NewPatternSet(),
		templateData:    make(map[string]interface{}),
		templateOptions: DefaultTemplateOptions,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// Add adds sourceStateEntry to s.
func (s *SourceState) Add() error {
	return nil // FIXME
}

// ApplyAll updates targetDir in fs to match s.
func (s *SourceState) ApplyAll(system System, umask os.FileMode, targetDir string) error {
	for _, targetName := range s.sortedTargetNames() {
		if err := s.ApplyOne(system, umask, targetDir, targetName); err != nil {
			return err
		}
	}
	return nil
}

// ApplyOne updates targetName in targetDir on fs to match s using s.
func (s *SourceState) ApplyOne(system System, umask os.FileMode, targetDir, targetName string) error {
	targetPath := path.Join(targetDir, targetName)
	destStateEntry, err := NewDestStateEntry(system, targetPath)
	if err != nil {
		return err
	}
	targetStateEntry, err := s.entries[targetName].TargetStateEntry()
	if err != nil {
		return err
	}
	if err := targetStateEntry.Apply(system, destStateEntry); err != nil {
		return err
	}
	if targetStateDir, ok := targetStateEntry.(*TargetStateDir); ok {
		if targetStateDir.exact {
			infos, err := system.ReadDir(targetPath)
			if err != nil {
				return err
			}
			baseNames := make([]string, 0, len(infos))
			for _, info := range infos {
				if baseName := info.Name(); baseName != "." && baseName != ".." {
					baseNames = append(baseNames, baseName)
				}
			}
			sort.Strings(baseNames)
			for _, baseName := range baseNames {
				if _, ok := s.entries[path.Join(targetName, baseName)]; !ok {
					if err := system.RemoveAll(path.Join(targetPath, baseName)); err != nil {
						return err
					}
				}
			}
		}
	}
	// FIXME chezmoiremove
	return nil
}

// Entries returns s's source state entries.
func (s *SourceState) Entries() map[string]SourceStateEntry {
	return s.entries
}

// TargetNames returns all of s's target names in alphabetical order.
func (s *SourceState) TargetNames() []string {
	targetNames := make([]string, 0, len(s.entries))
	for targetName := range s.entries {
		targetNames = append(targetNames, targetName)
	}
	sort.Strings(targetNames)
	return targetNames
}

// Entry returns the source state entry for targetName.
func (s *SourceState) Entry(targetName string) (SourceStateEntry, bool) {
	sourceStateEntry, ok := s.entries[targetName]
	return sourceStateEntry, ok
}

// Evaluate evaluates every target state entry in s.
func (s *SourceState) Evaluate() error {
	for _, targetName := range s.sortedTargetNames() {
		sourceStateEntry := s.entries[targetName]
		if err := sourceStateEntry.Evaluate(); err != nil {
			return err
		}
		targetStateEntry, err := sourceStateEntry.TargetStateEntry()
		if err != nil {
			return err
		}
		if err := targetStateEntry.Evaluate(); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteTemplateData returns the result of executing template data.
func (s *SourceState) ExecuteTemplateData(name string, data []byte) ([]byte, error) {
	tmpl, err := template.New(name).Option(s.templateOptions...).Funcs(s.templateFuncs).Parse(string(data))
	if err != nil {
		return nil, err
	}
	for name, t := range s.templates {
		tmpl, err = tmpl.AddParseTree(name, t.Tree)
		if err != nil {
			return nil, err
		}
	}
	output := &bytes.Buffer{}
	if err = tmpl.ExecuteTemplate(output, name, s.TemplateData()); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// MergeTemplateData merges templateData into s's template data.
func (s *SourceState) MergeTemplateData(templateData map[string]interface{}) {
	recursiveMerge(s.templateData, templateData)
}

// MustEntry returns the source state entry associated with targetName, and
// panics if it does not exist.
func (s *SourceState) MustEntry(targetName string) SourceStateEntry {
	sourceStateEntry, ok := s.entries[targetName]
	if !ok {
		panic(fmt.Sprintf("%s: no source state entry", targetName))
	}
	return sourceStateEntry
}

// Read reads a source state from sourcePath.
func (s *SourceState) Read() error {
	_, err := s.system.Stat(s.sourcePath)
	switch {
	case os.IsNotExist(err):
		return nil
	case err != nil:
		return err
	}

	// Read all source entries.
	allSourceStateEntries := make(map[string][]SourceStateEntry)
	sourceDirPrefix := filepath.ToSlash(s.sourcePath) + PathSeparatorStr
	if err := vfs.Walk(s.system, s.sourcePath, func(sourcePath string, info os.FileInfo, err error) error {
		sourcePath = filepath.ToSlash(sourcePath)
		if err != nil {
			return err
		}
		if sourcePath == s.sourcePath {
			return nil
		}
		relPath := strings.TrimPrefix(sourcePath, sourceDirPrefix)
		sourceDirName, sourceName := path.Split(relPath)
		targetDirName := getTargetDirName(sourceDirName)
		switch {
		case strings.HasPrefix(info.Name(), dataName):
			return s.addTemplateData(sourcePath)
		case info.Name() == ignoreName:
			return s.addPatterns(s.ignore, sourcePath, sourceDirName)
		case info.Name() == removeName:
			return s.addPatterns(s.remove, sourcePath, targetDirName)
		case info.Name() == templatesDirName:
			if err := s.addTemplatesDir(sourcePath); err != nil {
				return err
			}
			return filepath.SkipDir
		case info.Name() == versionName:
			return s.addVersionFile(sourcePath)
		case strings.HasPrefix(info.Name(), chezmoiPrefix):
			// FIXME accumulate warning about unrecognized special file
			fallthrough
		case strings.HasPrefix(info.Name(), ignorePrefix):
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		case info.IsDir():
			da := parseDirAttributes(sourceName)
			targetName := path.Join(targetDirName, da.Name)
			if s.ignore.Match(targetName) {
				return nil
			}
			sourceStateEntry := s.newSourceStateDir(sourcePath, da)
			allSourceStateEntries[targetName] = append(allSourceStateEntries[targetName], sourceStateEntry)
			return nil
		case info.Mode().IsRegular():
			fa := parseFileAttributes(sourceName)
			targetName := path.Join(targetDirName, fa.Name)
			if s.ignore.Match(targetName) {
				return nil
			}
			sourceStateEntry := s.newSourceStateFile(sourcePath, fa)
			allSourceStateEntries[targetName] = append(allSourceStateEntries[targetName], sourceStateEntry)
			return nil
		default:
			return &unsupportedFileTypeError{
				path: sourcePath,
				mode: info.Mode(),
			}
		}
	}); err != nil {
		return err
	}

	// Checking for duplicate source entries with the same target name. Iterate
	// over the target names in order so that any error is deterministic.
	targetNames := make([]string, 0, len(allSourceStateEntries))
	for targetName := range allSourceStateEntries {
		targetNames = append(targetNames, targetName)
	}
	sort.Strings(targetNames)
	for _, targetName := range targetNames {
		sourceStateEntries := allSourceStateEntries[targetName]
		if len(sourceStateEntries) == 1 {
			continue
		}
		sourcePaths := make([]string, 0, len(sourceStateEntries))
		for _, sourceStateEntry := range sourceStateEntries {
			sourcePaths = append(sourcePaths, sourceStateEntry.Path())
		}
		err = multierr.Append(err, &duplicateTargetError{
			targetName:  targetName,
			sourcePaths: sourcePaths,
		})
	}
	if err != nil {
		return err
	}

	// Populate s.Entries with the unique source entry for each target.
	for targetName, sourceEntries := range allSourceStateEntries {
		s.entries[targetName] = sourceEntries[0]
	}

	return nil
}

// Remove removes everything in targetDir that matches s's remove pattern set.
func (s *SourceState) Remove(system System, targetDir string) error {
	// Build a set of targets to remove.
	targetDirPrefix := targetDir + PathSeparatorStr
	targetPathsToRemove := newStringSet()
	for include := range s.remove.includes {
		matches, err := system.Glob(path.Join(targetDir, include))
		if err != nil {
			return err
		}
		for _, match := range matches {
			// Don't remove targets that are excluded from remove.
			if !s.remove.Match(strings.TrimPrefix(match, targetDirPrefix)) {
				continue
			}
			targetPathsToRemove.Add(match)
		}
	}

	// Remove targets in order. Parent directories are removed before their
	// children, which is okay because RemoveAll does not treat os.ErrNotExist
	// as an error.
	sortedTargetPathsToRemove := targetPathsToRemove.Elements()
	sort.Strings(sortedTargetPathsToRemove)
	for _, targetPath := range sortedTargetPathsToRemove {
		if err := system.RemoveAll(targetPath); err != nil {
			return err
		}
	}
	return nil
}

// TemplateData returns s's template data.
func (s *SourceState) TemplateData() map[string]interface{} {
	return s.templateData
}

func (s *SourceState) addPatterns(patternSet *PatternSet, sourcePath, relPath string) error {
	data, err := s.executeTemplate(sourcePath)
	if err != nil {
		return err
	}
	dir := filepath.Dir(relPath)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		text := scanner.Text()
		if index := strings.IndexRune(text, '#'); index != -1 {
			text = text[:index]
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		include := true
		if strings.HasPrefix(text, "!") {
			include = false
			text = strings.TrimPrefix(text, "!")
		}
		pattern := path.Join(dir, text)
		if err := patternSet.Add(pattern, include); err != nil {
			return fmt.Errorf("%s: %w", sourcePath, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s: %w", sourcePath, err)
	}
	return nil
}

func (s *SourceState) addTemplateData(sourcePath string) error {
	_, name := path.Split(sourcePath)
	suffix := strings.TrimPrefix(name, dataName+".")
	format, ok := Formats[strings.ToLower(suffix)]
	if !ok {
		return fmt.Errorf("%s: unknown format", sourcePath)
	}
	data, err := s.system.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("%s: %w", sourcePath, err)
	}
	var templateData map[string]interface{}
	if err := format.Decode(data, &templateData); err != nil {
		return fmt.Errorf("%s: %w", sourcePath, err)
	}
	recursiveMerge(s.templateData, templateData)
	return nil
}

func (s *SourceState) addTemplatesDir(templateDir string) error {
	templateDirPrefix := filepath.ToSlash(templateDir) + PathSeparatorStr
	return vfs.Walk(s.system, templateDir, func(templatePath string, info os.FileInfo, err error) error {
		templatePath = filepath.ToSlash(templatePath)
		if err != nil {
			return err
		}
		switch {
		case info.Mode().IsRegular():
			contents, err := s.system.ReadFile(templatePath)
			if err != nil {
				return err
			}
			name := strings.TrimPrefix(templatePath, templateDirPrefix)
			tmpl, err := template.New(name).Parse(string(contents))
			if err != nil {
				return err
			}
			if s.templates == nil {
				s.templates = make(map[string]*template.Template)
			}
			s.templates[name] = tmpl
			return nil
		case info.IsDir():
			return nil
		default:
			return &unsupportedFileTypeError{
				path: templatePath,
				mode: info.Mode(),
			}
		}
	})
}

// addVersionFile reads a .chezmoiversion file from source path and updates s's
// minimum version if it contains a more recent version than the current minimum
// version.
func (s *SourceState) addVersionFile(sourcePath string) error {
	data, err := s.system.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	version, err := semver.NewVersion(strings.TrimSpace(string(data)))
	if err != nil {
		return err
	}
	if s.minVersion == nil || s.minVersion.LessThan(*version) {
		s.minVersion = version
	}
	return nil
}

func (s *SourceState) executeTemplate(path string) ([]byte, error) {
	data, err := s.system.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return s.ExecuteTemplateData(path, data)
}

func (s *SourceState) newSourceStateDir(sourcePath string, da dirAttributes) *SourceStateDir {
	perm := os.FileMode(0o777)
	if da.Private {
		perm &^= 0o77
	}
	perm &^= s.umask

	targetStateDir := &TargetStateDir{
		perm:  perm,
		exact: da.Exact,
	}

	return &SourceStateDir{
		path:             sourcePath,
		attributes:       da,
		targetStateEntry: targetStateDir,
	}
}

func (s *SourceState) newSourceStateFile(sourcePath string, fa fileAttributes) *SourceStateFile {
	lazyContents := &lazyContents{
		contentsFunc: func() ([]byte, error) {
			contents, err := s.system.ReadFile(sourcePath)
			if err != nil {
				return nil, err
			}
			if !fa.Encrypted {
				return contents, nil
			}
			// FIXME pass targetName as filenameHint
			return s.encryptionTool.Decrypt(sourcePath, contents)
		},
	}

	var targetStateEntryFunc func() (TargetStateEntry, error)
	switch fa.Type {
	case sourceFileTypeFile:
		targetStateEntryFunc = func() (TargetStateEntry, error) {
			contents, err := lazyContents.Contents()
			if err != nil {
				return nil, err
			}
			if fa.Template {
				contents, err = s.ExecuteTemplateData(sourcePath, contents)
				if err != nil {
					return nil, err
				}
			}
			if !fa.Empty && isEmpty(contents) {
				return &TargetStateAbsent{}, nil
			}
			perm := os.FileMode(0o666)
			if fa.Executable {
				perm |= 0o111
			}
			if fa.Private {
				perm &^= 0o77
			}
			perm &^= s.umask
			return &TargetStateFile{
				lazyContents: newLazyContents(contents),
				perm:         perm,
			}, nil
		}
	case sourceFileTypeScript:
		targetStateEntryFunc = func() (TargetStateEntry, error) {
			contents, err := lazyContents.Contents()
			if err != nil {
				return nil, err
			}
			if fa.Template {
				contents, err = s.ExecuteTemplateData(sourcePath, contents)
				if err != nil {
					return nil, err
				}
			}
			return &TargetStateScript{
				lazyContents: newLazyContents(contents),
				name:         fa.Name,
				once:         fa.Once,
			}, nil
		}
	case sourceFileTypeSymlink:
		targetStateEntryFunc = func() (TargetStateEntry, error) {
			linknameBytes, err := lazyContents.Contents()
			if err != nil {
				return nil, err
			}
			if fa.Template {
				linknameBytes, err = s.ExecuteTemplateData(sourcePath, linknameBytes)
				if err != nil {
					return nil, err
				}
			}
			return &TargetStateSymlink{
				lazyLinkname: newLazyLinkname(string(bytes.TrimSpace(linknameBytes))),
			}, nil
		}
	default:
		panic(fmt.Sprintf("unsupported type: %s", string(fa.Type)))
	}

	return &SourceStateFile{
		lazyContents:         lazyContents,
		path:                 sourcePath,
		attributes:           fa,
		targetStateEntryFunc: targetStateEntryFunc,
	}
}

// sortedTargetNames returns all of s's target names in order.
func (s *SourceState) sortedTargetNames() []string {
	targetNames := make([]string, 0, len(s.entries))
	for targetName := range s.entries {
		targetNames = append(targetNames, targetName)
	}
	sort.Strings(targetNames)
	return targetNames
}

// getTargetDirName returns the target directory name of sourceDirName.
func getTargetDirName(sourceDirName string) string {
	sourceNames := strings.Split(sourceDirName, PathSeparatorStr)
	targetNames := make([]string, 0, len(sourceNames))
	for _, sourceName := range sourceNames {
		da := parseDirAttributes(sourceName)
		targetNames = append(targetNames, da.Name)
	}
	return strings.Join(targetNames, PathSeparatorStr)
}
