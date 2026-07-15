//go:build windows

package historyui

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"

	"whispr/internal/history"
)

const (
	windowClass = "WhisprHistory"

	idList    = 1001
	idCopy    = 1002
	idRefresh = 1003

	wmClose       = 0x0010
	wmCommand     = 0x0111
	wmDestroy     = 0x0002
	wmNotify      = 0x004E
	wmSetFont     = 0x0030
	wsCaption     = 0x00C00000
	wsChild       = 0x40000000
	wsDisabled    = 0x08000000
	wsMinimizeBox = 0x00020000
	wsSysMenu     = 0x00080000
	wsVisible     = 0x10000000
	wsVScroll     = 0x00200000

	wsExClientEdge = 0x00000200

	cwUseDefault = 0x80000000
	swShow       = 5

	bsPushButton  = 0x00000000
	esAutoVScroll = 0x0040
	esMultiline   = 0x0004
	esReadOnly    = 0x0800

	lvsReport        = 0x0001
	lvsSingleSel     = 0x0004
	lvsShowSelAlways = 0x0008
	lvsExFullRowSel  = 0x00000020

	lvmFirst                    = 0x1000
	lvmDeleteAllItems           = lvmFirst + 9
	lvmGetNextItem              = lvmFirst + 12
	lvmSetExtendedListViewStyle = lvmFirst + 54
	lvmInsertItemW              = lvmFirst + 77
	lvmInsertColumnW            = lvmFirst + 97
	lvmSetItemTextW             = lvmFirst + 116

	lvifText     = 0x0001
	lvniSelected = 0x0002
	lvcfFmt      = 0x0001
	lvcfWidth    = 0x0002
	lvcfText     = 0x0004
	lvcfSubItem  = 0x0008
	lvcfmtLeft   = 0

	iccListViewClasses = 0x00000001

	bnClicked      = 0
	nmDblClk       = -3
	lvnItemChanged = -101

	cfUnicodeText = 13
	gmemMoveable  = 0x0002
	gmemZeroInit  = 0x0040

	defaultGuiFont = 17
)

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type lvColumn struct {
	Mask       uint32
	Fmt        int32
	Cx         int32
	Text       *uint16
	TextMax    int32
	SubItem    int32
	Image      int32
	Order      int32
	CxMin      int32
	CxDefault  int32
	IdealWidth int32
}

type lvItem struct {
	Mask       uint32
	Item       int32
	SubItem    int32
	State      uint32
	StateMask  uint32
	Text       *uint16
	TextMax    int32
	Image      int32
	Param      uintptr
	Indent     int32
	GroupID    int32
	Columns    uint32
	ColumnsPtr uintptr
}

type nmhdr struct {
	HwndFrom uintptr
	IDFrom   uintptr
	Code     int32
}

type initCommonControlsEx struct {
	Size uint32
	ICC  uint32
}

type window struct {
	hwnd     uintptr
	list     uintptr
	detail   uintptr
	status   uintptr
	copyBtn  uintptr
	refBtn   uintptr
	path     string
	entries  []history.Entry
	selected int
}

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	comctl32 = windows.NewLazySystemDLL("comctl32.dll")

	getModuleHandle     = kernel32.NewProc("GetModuleHandleW")
	globalAlloc         = kernel32.NewProc("GlobalAlloc")
	globalFree          = kernel32.NewProc("GlobalFree")
	globalLock          = kernel32.NewProc("GlobalLock")
	globalUnlock        = kernel32.NewProc("GlobalUnlock")
	registerClassEx     = user32.NewProc("RegisterClassExW")
	createWindowEx      = user32.NewProc("CreateWindowExW")
	defWindowProc       = user32.NewProc("DefWindowProcW")
	destroyWindow       = user32.NewProc("DestroyWindow")
	showWindow          = user32.NewProc("ShowWindow")
	updateWindow        = user32.NewProc("UpdateWindow")
	getMessage          = user32.NewProc("GetMessageW")
	translateMessage    = user32.NewProc("TranslateMessage")
	dispatchMessage     = user32.NewProc("DispatchMessageW")
	postQuitMessage     = user32.NewProc("PostQuitMessage")
	sendMessage         = user32.NewProc("SendMessageW")
	setWindowText       = user32.NewProc("SetWindowTextW")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
	openClipboard       = user32.NewProc("OpenClipboard")
	closeClipboard      = user32.NewProc("CloseClipboard")
	emptyClipboard      = user32.NewProc("EmptyClipboard")
	setClipboardData    = user32.NewProc("SetClipboardData")
	getStockObject      = gdi32.NewProc("GetStockObject")
	initCommonControls  = comctl32.NewProc("InitCommonControlsEx")

	className = utf16(windowClass)
	regOnce   sync.Once
	openMu    sync.Mutex
	isOpen    bool
	active    uintptr
	winMu     sync.Mutex
	wins      = map[uintptr]*window{}
)

func Open(path string) error {
	openMu.Lock()
	if isOpen {
		hwnd := active
		openMu.Unlock()
		if hwnd != 0 {
			showWindow.Call(hwnd, swShow)
			setForegroundWindow.Call(hwnd)
		}
		return nil
	}
	isOpen = true
	ready := make(chan error, 1)
	go func() {
		_ = run(path, ready)
		openMu.Lock()
		isOpen = false
		active = 0
		openMu.Unlock()
	}()
	openMu.Unlock()

	if err := <-ready; err != nil {
		return err
	}
	return nil
}

func run(path string, ready chan<- error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var icc = initCommonControlsEx{Size: uint32(unsafe.Sizeof(initCommonControlsEx{})), ICC: iccListViewClasses}
	initCommonControls.Call(uintptr(unsafe.Pointer(&icc)))

	hinst, _, _ := getModuleHandle.Call(0)
	regOnce.Do(func() {
		wc := wndClassEx{Size: uint32(unsafe.Sizeof(wndClassEx{})), WndProc: windows.NewCallback(wndProc), Instance: hinst, ClassName: className}
		registerClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	})

	hwnd, _, err := createWindowEx.Call(0, ptr(windowClass), ptr("Whispr History"), wsCaption|wsSysMenu|wsMinimizeBox, cwUseDefault, cwUseDefault, 900, 560, 0, 0, hinst, 0)
	if hwnd == 0 {
		ready <- fmt.Errorf("create history window: %w", err)
		return err
	}
	w := &window{hwnd: hwnd, path: path, selected: -1}
	setWindow(hwnd, w)
	defer delWindow(hwnd)

	if err := w.createControls(hinst); err != nil {
		ready <- err
		destroyWindow.Call(hwnd)
		return err
	}
	w.reload()
	showWindow.Call(hwnd, swShow)
	updateWindow.Call(hwnd)

	openMu.Lock()
	active = hwnd
	openMu.Unlock()
	ready <- nil

	var m msg
	for {
		r, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if r == 0 || r == ^uintptr(0) {
			return nil
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&m)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func (w *window) createControls(hinst uintptr) error {
	font, _, _ := getStockObject.Call(defaultGuiFont)
	w.list = createControl(wsExClientEdge, "SysListView32", "", wsChild|wsVisible|lvsReport|lvsSingleSel|lvsShowSelAlways, 12, 12, 860, 290, w.hwnd, idList, hinst)
	w.detail = createControl(wsExClientEdge, "EDIT", "", wsChild|wsVisible|wsVScroll|esMultiline|esReadOnly|esAutoVScroll, 12, 312, 860, 145, w.hwnd, 0, hinst)
	w.copyBtn = createControl(0, "BUTTON", "Copy Cleaned", wsChild|wsVisible|bsPushButton, 12, 468, 130, 32, w.hwnd, idCopy, hinst)
	w.refBtn = createControl(0, "BUTTON", "Refresh", wsChild|wsVisible|bsPushButton, 154, 468, 100, 32, w.hwnd, idRefresh, hinst)
	w.status = createControl(0, "STATIC", "", wsChild|wsVisible, 270, 474, 600, 24, w.hwnd, 0, hinst)
	if w.list == 0 || w.detail == 0 || w.copyBtn == 0 || w.refBtn == 0 || w.status == 0 {
		return fmt.Errorf("create history controls")
	}
	for _, hwnd := range []uintptr{w.list, w.detail, w.copyBtn, w.refBtn, w.status} {
		sendMessage.Call(hwnd, wmSetFont, font, 1)
	}
	sendMessage.Call(w.list, lvmSetExtendedListViewStyle, 0, lvsExFullRowSel)
	w.insertColumn(0, "Time", 150)
	w.insertColumn(1, "Raw", 320)
	w.insertColumn(2, "Cleaned", 370)
	return nil
}

func createControl(ex uintptr, class, title string, style uintptr, x, y, w, h int, parent uintptr, id int, hinst uintptr) uintptr {
	hwnd, _, _ := createWindowEx.Call(ex, ptr(class), ptr(title), style, uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, uintptr(id), hinst, 0)
	return hwnd
}

func (w *window) insertColumn(i int, title string, width int) {
	col := lvColumn{Mask: lvcfFmt | lvcfWidth | lvcfText | lvcfSubItem, Fmt: lvcfmtLeft, Cx: int32(width), Text: utf16(title), SubItem: int32(i)}
	sendMessage.Call(w.list, lvmInsertColumnW, uintptr(i), uintptr(unsafe.Pointer(&col)))
}

func (w *window) reload() {
	got, err := history.LoadLatest(w.path, history.DefaultLimit)
	if err != nil {
		w.setStatus(err.Error())
		return
	}
	w.entries = got.Entries
	w.selected = -1
	setWindowText.Call(w.detail, ptr(""))
	sendMessage.Call(w.list, lvmDeleteAllItems, 0, 0)
	for i, e := range w.entries {
		w.insertRow(i, e)
	}
	switch {
	case len(w.entries) == 0 && got.Skipped > 0:
		w.setStatus(fmt.Sprintf("No history yet · skipped %d invalid", got.Skipped))
	case len(w.entries) == 0:
		w.setStatus("No history yet")
	case got.Skipped > 0:
		w.setStatus(fmt.Sprintf("Loaded %d entries · skipped %d invalid", len(w.entries), got.Skipped))
	default:
		w.setStatus(fmt.Sprintf("Loaded %d entries", len(w.entries)))
	}
}

func (w *window) insertRow(i int, e history.Entry) {
	item := lvItem{Mask: lvifText, Item: int32(i), Text: utf16(e.Time)}
	sendMessage.Call(w.list, lvmInsertItemW, 0, uintptr(unsafe.Pointer(&item)))
	w.setCell(i, 1, oneLine(e.Raw))
	w.setCell(i, 2, oneLine(e.Cleaned))
}

func (w *window) setCell(row, col int, text string) {
	item := lvItem{Item: int32(row), SubItem: int32(col), Text: utf16(text)}
	sendMessage.Call(w.list, lvmSetItemTextW, uintptr(row), uintptr(unsafe.Pointer(&item)))
}

func (w *window) updateSelection() {
	w.selected = w.selectedIndex()
	if w.selected < 0 || w.selected >= len(w.entries) {
		setWindowText.Call(w.detail, ptr(""))
		return
	}
	setWindowText.Call(w.detail, ptr(w.entries[w.selected].Cleaned))
}

func (w *window) selectedIndex() int {
	r, _, _ := sendMessage.Call(w.list, lvmGetNextItem, ^uintptr(0), lvniSelected)
	if r == ^uintptr(0) {
		return -1
	}
	return int(r)
}

func (w *window) copySelected() {
	w.updateSelection()
	if w.selected < 0 || w.selected >= len(w.entries) {
		w.setStatus("Select history first")
		return
	}
	if err := setClipboardText(w.hwnd, w.entries[w.selected].Cleaned); err != nil {
		w.setStatus(err.Error())
		return
	}
	w.setStatus("Copied")
}

func (w *window) setStatus(s string) {
	setWindowText.Call(w.status, ptr(s))
}

func wndProc(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	w := getWindow(hwnd)
	switch msg {
	case wmCommand:
		if w != nil && hiword(wparam) == bnClicked {
			switch loword(wparam) {
			case idCopy:
				w.copySelected()
				return 0
			case idRefresh:
				w.reload()
				return 0
			}
		}
	case wmNotify:
		if w != nil {
			nm := (*nmhdr)(unsafe.Pointer(lparam))
			if nm.IDFrom == idList {
				switch nm.Code {
				case lvnItemChanged:
					w.updateSelection()
				case nmDblClk:
					w.copySelected()
				}
			}
		}
	case wmClose:
		destroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		postQuitMessage.Call(0)
		return 0
	}
	r, _, _ := defWindowProc.Call(hwnd, uintptr(msg), wparam, lparam)
	return r
}

func setClipboardText(hwnd uintptr, s string) error {
	if r, _, err := openClipboard.Call(hwnd); r == 0 {
		return fmt.Errorf("open clipboard: %w", err)
	}
	defer closeClipboard.Call()
	emptyClipboard.Call()

	data := windows.StringToUTF16(s)
	h, _, err := globalAlloc.Call(gmemMoveable|gmemZeroInit, uintptr(len(data)*2))
	if h == 0 {
		return fmt.Errorf("allocate clipboard: %w", err)
	}
	p, _, err := globalLock.Call(h)
	if p == 0 {
		globalFree.Call(h)
		return fmt.Errorf("lock clipboard: %w", err)
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(data)), data)
	globalUnlock.Call(h)
	if r, _, err := setClipboardData.Call(cfUnicodeText, h); r == 0 {
		globalFree.Call(h)
		return fmt.Errorf("set clipboard: %w", err)
	}
	return nil
}

func setWindow(hwnd uintptr, w *window) {
	winMu.Lock()
	wins[hwnd] = w
	winMu.Unlock()
}

func getWindow(hwnd uintptr) *window {
	winMu.Lock()
	w := wins[hwnd]
	winMu.Unlock()
	return w
}

func delWindow(hwnd uintptr) {
	winMu.Lock()
	delete(wins, hwnd)
	winMu.Unlock()
}

func ptr(s string) uintptr { return uintptr(unsafe.Pointer(utf16(s))) }

func utf16(s string) *uint16 {
	p, err := windows.UTF16PtrFromString(s)
	if err == nil {
		return p
	}
	p, _ = windows.UTF16PtrFromString(strings.ReplaceAll(s, "\x00", " "))
	return p
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(s, "\t", " ")
}

func loword(v uintptr) int { return int(v & 0xffff) }
func hiword(v uintptr) int { return int((v >> 16) & 0xffff) }
