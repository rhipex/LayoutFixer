//go:build windows

package main

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

const (
	WM_DESTROY                = 0x0002
	WM_CLOSE                  = 0x0010
	WM_INPUTLANGCHANGEREQUEST = 0x0050
	WM_COMMAND                = 0x0111
	WM_RBUTTONUP              = 0x0205
	WM_LBUTTONUP              = 0x0202
	WM_APP                    = 0x8000
	WM_TRAY                   = WM_APP + 1
	WM_CONVERT                = WM_APP + 2

	WH_KEYBOARD_LL = 13
	HC_ACTION      = 0
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105
	LLKHF_INJECTED = 0x00000010

	VK_LWIN     = 0x5B
	VK_RWIN     = 0x5C
	VK_LSHIFT   = 0xA0
	VK_RSHIFT   = 0xA1
	VK_SHIFT    = 0x10
	VK_CONTROL  = 0x11
	VK_LCONTROL = 0xA2
	VK_RCONTROL = 0xA3
	VK_MENU     = 0x12
	VK_C        = 0x43
	VK_V        = 0x56

	INPUT_KEYBOARD        = 1
	KEYEVENTF_EXTENDEDKEY = 0x0001
	KEYEVENTF_KEYUP       = 0x0002

	NIM_ADD     = 0x00000000
	NIM_DELETE  = 0x00000002
	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	MF_STRING       = 0x00000000
	MF_SEPARATOR    = 0x00000800
	MF_CHECKED      = 0x00000008
	MF_GRAYED       = 0x00000001
	TPM_RIGHTBUTTON = 0x0002
	TPM_RETURNCMD   = 0x0100

	ID_TRAY    = 1
	ID_STARTUP = 1001
	ID_EXIT    = 1002

	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002

	KEY_QUERY_VALUE      = 0x0001
	KEY_SET_VALUE        = 0x0002
	REG_SZ               = 1
	ERROR_SUCCESS        = 0
	ERROR_FILE_NOT_FOUND = 2
	ERROR_ALREADY_EXISTS = 183

	HKEY_CURRENT_USER = 0x80000001

	LANG_RUSSIAN = 0x19

	LR_DEFAULTCOLOR = 0x0000
	ICON_VERSION    = 0x00030000
	VK_MENU_MASK    = 0xE8
	IDI_APPLICATION = 32512

	MB_OK              = 0x00000000
	MB_ICONINFORMATION = 0x00000040
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")

	pRegisterClassW             = user32.NewProc("RegisterClassW")
	pCreateWindowExW            = user32.NewProc("CreateWindowExW")
	pDefWindowProcW             = user32.NewProc("DefWindowProcW")
	pDestroyWindow              = user32.NewProc("DestroyWindow")
	pPostQuitMessage            = user32.NewProc("PostQuitMessage")
	pGetMessageW                = user32.NewProc("GetMessageW")
	pTranslateMessage           = user32.NewProc("TranslateMessage")
	pDispatchMessageW           = user32.NewProc("DispatchMessageW")
	pSetWindowsHookExW          = user32.NewProc("SetWindowsHookExW")
	pUnhookWindowsHookEx        = user32.NewProc("UnhookWindowsHookEx")
	pCallNextHookEx             = user32.NewProc("CallNextHookEx")
	pPostMessageW               = user32.NewProc("PostMessageW")
	pGetForegroundWindow        = user32.NewProc("GetForegroundWindow")
	pGetWindowThreadProcessId   = user32.NewProc("GetWindowThreadProcessId")
	pGetKeyboardLayout          = user32.NewProc("GetKeyboardLayout")
	pGetAsyncKeyState           = user32.NewProc("GetAsyncKeyState")
	pLoadKeyboardLayoutW        = user32.NewProc("LoadKeyboardLayoutW")
	pSendInput                  = user32.NewProc("SendInput")
	pOpenClipboard              = user32.NewProc("OpenClipboard")
	pCloseClipboard             = user32.NewProc("CloseClipboard")
	pGetClipboardData           = user32.NewProc("GetClipboardData")
	pEmptyClipboard             = user32.NewProc("EmptyClipboard")
	pSetClipboardData           = user32.NewProc("SetClipboardData")
	pGetClipboardSequenceNumber = user32.NewProc("GetClipboardSequenceNumber")
	pCreatePopupMenu            = user32.NewProc("CreatePopupMenu")
	pAppendMenuW                = user32.NewProc("AppendMenuW")
	pTrackPopupMenu             = user32.NewProc("TrackPopupMenu")
	pDestroyMenu                = user32.NewProc("DestroyMenu")
	pGetCursorPos               = user32.NewProc("GetCursorPos")
	pSetForegroundWindow        = user32.NewProc("SetForegroundWindow")
	pLoadIconW                  = user32.NewProc("LoadIconW")
	pCreateIconFromResourceEx   = user32.NewProc("CreateIconFromResourceEx")
	pDestroyIcon                = user32.NewProc("DestroyIcon")
	pMessageBoxW                = user32.NewProc("MessageBoxW")

	pShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")

	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	pGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	pGlobalLock       = kernel32.NewProc("GlobalLock")
	pGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	pGlobalFree       = kernel32.NewProc("GlobalFree")
	pCreateMutexW     = kernel32.NewProc("CreateMutexW")
	pGetLastError     = kernel32.NewProc("GetLastError")

	pRegOpenKeyExW    = advapi32.NewProc("RegOpenKeyExW")
	pRegCreateKeyExW  = advapi32.NewProc("RegCreateKeyExW")
	pRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	pRegSetValueExW   = advapi32.NewProc("RegSetValueExW")
	pRegDeleteValueW  = advapi32.NewProc("RegDeleteValueW")
	pRegCloseKey      = advapi32.NewProc("RegCloseKey")

	pOleInitialize   = ole32.NewProc("OleInitialize")
	pOleUninitialize = ole32.NewProc("OleUninitialize")
	pOleGetClipboard = ole32.NewProc("OleGetClipboard")
	pOleSetClipboard = ole32.NewProc("OleSetClipboard")
)

type WNDCLASS struct {
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
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
	Private uint32
}

type POINT struct {
	X int32
	Y int32
}

type KBDLLHOOKSTRUCT struct {
	VkCode    uint32
	ScanCode  uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type KEYBDINPUT struct {
	Vk        uint16
	Scan      uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type INPUT struct {
	Type    uint32
	Padding uint32
	Ki      KEYBDINPUT
	Extra   [8]byte
}

type NOTIFYICONDATA struct {
	CbSize            uint32
	HWnd              uintptr
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             uintptr
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          [16]byte
	HBalloonIcon      uintptr
}

const embeddedIconBase64 = "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"

var embeddedIcon = mustDecodeBase64(embeddedIconBase64)

var (
	hwnd       uintptr
	hookHandle uintptr
	appIcon    uintptr

	keyDown [256]bool

	chordActive   bool
	chordFirstVK  uint32 // modifier key-down that Windows already received
	chordSecondVK uint32 // modifier key-down suppressed by our hook

	conversionBusy int32
)

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func copyUTF16(dst []uint16, s string) {
	u := utf16.Encode([]rune(s + "\x00"))
	copy(dst, u)
}

func mustDecodeBase64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func callNext(nCode int, wParam, lParam uintptr) uintptr {
	r, _, _ := pCallNextHookEx.Call(hookHandle, uintptr(nCode), wParam, lParam)
	return r
}

func isWin(vk uint32) bool   { return vk == VK_LWIN || vk == VK_RWIN }
func isShift(vk uint32) bool { return vk == VK_LSHIFT || vk == VK_RSHIFT || vk == VK_SHIFT }

func anyWinDown() bool {
	return keyDown[VK_LWIN] || keyDown[VK_RWIN]
}

func anyShiftDown() bool {
	return keyDown[VK_LSHIFT] || keyDown[VK_RSHIFT] || keyDown[VK_SHIFT]
}

func rememberPhysical(vk uint32, down bool) {
	if vk < uint32(len(keyDown)) {
		keyDown[vk] = down
	}
}

func downWinVK() uint32 {
	if keyDown[VK_LWIN] {
		return VK_LWIN
	}
	if keyDown[VK_RWIN] {
		return VK_RWIN
	}
	return 0
}

func downShiftVK() uint32 {
	if keyDown[VK_LSHIFT] {
		return VK_LSHIFT
	}
	if keyDown[VK_RSHIFT] {
		return VK_RSHIFT
	}
	if keyDown[VK_SHIFT] {
		return VK_SHIFT
	}
	return 0
}

func keyInput(vk uint32, up bool, extended bool) INPUT {
	flags := uint32(0)
	if up {
		flags |= KEYEVENTF_KEYUP
	}
	if extended || isWin(vk) {
		flags |= KEYEVENTF_EXTENDEDKEY
	}
	return INPUT{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: uint16(vk), Flags: flags}}
}

func sendInputs(inputs []INPUT) {
	if len(inputs) == 0 {
		return
	}
	pSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(inputs[0]))
}

// The second modifier of Win+Shift is initially hidden from Windows. If another
// key follows (Win+Shift+S, etc.), recreate the hidden modifier followed by that
// key so normal Windows/app shortcuts still receive the original combination.
func replayCancelledChord(k *KBDLLHOOKSTRUCT) {
	extended := (k.Flags & 0x01) != 0 // LLKHF_EXTENDED
	sendInputs([]INPUT{
		keyInput(chordSecondVK, false, isWin(chordSecondVK)),
		keyInput(k.VkCode, false, extended),
	})
}

// If Win was the first modifier, Windows has already seen Win-down. A dummy
// unassigned key between Win-down and Win-up prevents the Start menu from
// opening when we consume the actual Win+Shift chord. This is the same class of
// workaround used by mature keyboard remappers for modifier-only hotkeys.
func releaseVisibleFirstModifier() {
	sendInputs([]INPUT{
		keyInput(VK_MENU_MASK, false, false),
		keyInput(VK_MENU_MASK, true, false),
		keyInput(chordFirstVK, true, isWin(chordFirstVK)),
	})
}

func finishChordIfReleased() {
	if chordActive && !anyWinDown() && !anyShiftDown() {
		chordActive = false
		chordFirstVK = 0
		chordSecondVK = 0
		pPostMessageW.Call(hwnd, WM_CONVERT, 0, 0)
	}
}

func keyboardProc(nCode int, wParam, lParam uintptr) uintptr {
	if nCode != HC_ACTION {
		return callNext(nCode, wParam, lParam)
	}
	k := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
	if (k.Flags & LLKHF_INJECTED) != 0 {
		return callNext(nCode, wParam, lParam)
	}

	down := wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN
	up := wParam == WM_KEYUP || wParam == WM_SYSKEYUP
	vk := k.VkCode

	if down {
		// Once the two-modifier candidate is active, any third key means the
		// user intended a regular Win+Shift+... shortcut. Recreate the hidden
		// second modifier and this key, then leave the rest of the sequence alone.
		if chordActive && vk != chordFirstVK && vk != chordSecondVK {
			// If the hidden second modifier is still physically held, rebuild the
			// intended Win+Shift+key shortcut. If it was already released, the
			// Win+Shift gesture was complete; cancel conversion and let this new
			// key behave with whichever modifier is actually still held.
			secondStillDown := chordSecondVK < uint32(len(keyDown)) && keyDown[chordSecondVK]
			if secondStillDown {
				replayCancelledChord(k)
				chordActive = false
				chordFirstVK = 0
				chordSecondVK = 0
				return 1 // physical key-down is replaced by the injected copy
			}
			chordActive = false
			chordFirstVK = 0
			chordSecondVK = 0
		}

		rememberPhysical(vk, true)

		if !chordActive && ((isWin(vk) && anyShiftDown()) || (isShift(vk) && anyWinDown())) {
			chordActive = true
			chordSecondVK = vk
			if isWin(vk) {
				chordFirstVK = downShiftVK()
			} else {
				chordFirstVK = downWinVK()
			}
			// Crucial difference from v1: Windows/the foreground app never sees
			// the second modifier for a modifier-only Win+Shift press.
			return 1
		}

		return callNext(nCode, wParam, lParam)
	}

	if up {
		rememberPhysical(vk, false)

		if chordActive {
			if vk == chordSecondVK {
				// Its key-down was suppressed, so suppress its key-up too.
				finishChordIfReleased()
				return 1
			}
			if vk == chordFirstVK {
				// Its key-down reached Windows. Replace the physical key-up with
				// a masked injected key-up to avoid Win/Alt-style menu side effects.
				releaseVisibleFirstModifier()
				finishChordIfReleased()
				return 1
			}
		}
	}

	return callNext(nCode, wParam, lParam)
}

func keyAsyncDown(vk uint32) bool {
	r, _, _ := pGetAsyncKeyState.Call(uintptr(vk))
	return (uint16(r) & 0x8000) != 0
}

// The conversion uses Ctrl+C/Ctrl+V internally. On some apps (notably
// Chromium/Firefox), Ctrl+Shift+C opens DevTools. A modifier-only hotkey can
// leave Shift logically down for a few milliseconds even after its physical
// release. Normalize Win/Shift state and wait for the physical keys to be up
// before issuing Ctrl+C.
func normalizeModifiersBeforeCopy() {
	deadline := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(deadline) {
		if !keyAsyncDown(VK_LSHIFT) && !keyAsyncDown(VK_RSHIFT) &&
			!keyAsyncDown(VK_LWIN) && !keyAsyncDown(VK_RWIN) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Physical state is now up. Force the logical state up too. The dummy key
	// prevents a synthetic Win-up from opening the Start menu.
	sendInputs([]INPUT{
		keyInput(VK_MENU_MASK, false, false),
		keyInput(VK_LSHIFT, true, false),
		keyInput(VK_RSHIFT, true, false),
		keyInput(VK_LWIN, true, true),
		keyInput(VK_RWIN, true, true),
		keyInput(VK_MENU_MASK, true, false),
	})
	time.Sleep(30 * time.Millisecond)
}

func sendCtrlKey(vk uint16) {
	inputs := [4]INPUT{
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: VK_CONTROL}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: vk}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: vk, Flags: KEYEVENTF_KEYUP}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{Vk: VK_CONTROL, Flags: KEYEVENTF_KEYUP}},
	}
	pSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(inputs[0]))
}

func getClipboardText() (string, bool) {
	for i := 0; i < 15; i++ {
		r, _, _ := pOpenClipboard.Call(hwnd)
		if r != 0 {
			defer pCloseClipboard.Call()
			h, _, _ := pGetClipboardData.Call(CF_UNICODETEXT)
			if h == 0 {
				return "", false
			}
			p, _, _ := pGlobalLock.Call(h)
			if p == 0 {
				return "", false
			}
			defer pGlobalUnlock.Call(h)
			chars := make([]uint16, 0, 256)
			for off := uintptr(0); ; off += 2 {
				c := *(*uint16)(unsafe.Pointer(p + off))
				if c == 0 {
					break
				}
				chars = append(chars, c)
			}
			return string(utf16.Decode(chars)), true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return "", false
}

func setClipboardText(s string) bool {
	u := utf16.Encode([]rune(s + "\x00"))
	size := uintptr(len(u) * 2)
	h, _, _ := pGlobalAlloc.Call(GMEM_MOVEABLE, size)
	if h == 0 {
		return false
	}
	p, _, _ := pGlobalLock.Call(h)
	if p == 0 {
		pGlobalFree.Call(h)
		return false
	}
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(u))
	copy(dst, u)
	pGlobalUnlock.Call(h)

	opened := false
	for i := 0; i < 15; i++ {
		r, _, _ := pOpenClipboard.Call(hwnd)
		if r != 0 {
			opened = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !opened {
		pGlobalFree.Call(h)
		return false
	}
	pEmptyClipboard.Call()
	r, _, _ := pSetClipboardData.Call(CF_UNICODETEXT, h)
	pCloseClipboard.Call()
	if r == 0 {
		pGlobalFree.Call(h)
		return false
	}
	// Ownership of h was transferred to the system.
	return true
}

func comRelease(p uintptr) {
	if p == 0 {
		return
	}
	vtbl := *(*uintptr)(unsafe.Pointer(p))
	release := *(*uintptr)(unsafe.Pointer(vtbl + 2*unsafe.Sizeof(uintptr(0))))
	syscall.SyscallN(release, p)
}

func saveClipboardObject() uintptr {
	var p uintptr
	hr, _, _ := pOleGetClipboard.Call(uintptr(unsafe.Pointer(&p)))
	if int32(hr) < 0 {
		return 0
	}
	return p
}

func restoreClipboardObject(p uintptr) {
	if p == 0 {
		return
	}
	pOleSetClipboard.Call(p)
	comRelease(p)
}

func primaryLangID(langID uint16) uint16 { return langID & 0x03ff }

type conversionDirection int

const (
	directionUnknown conversionDirection = iota
	directionFromRussian
	directionFromEnglish
)

var enToRuMap = map[rune]rune{
	'`': 'ё', '~': 'Ё', '@': '"', '#': '№', '$': ';', '^': ':', '&': '?',
	'q': 'й', 'w': 'ц', 'e': 'у', 'r': 'к', 't': 'е', 'y': 'н', 'u': 'г', 'i': 'ш', 'o': 'щ', 'p': 'з', '[': 'х', ']': 'ъ',
	'a': 'ф', 's': 'ы', 'd': 'в', 'f': 'а', 'g': 'п', 'h': 'р', 'j': 'о', 'k': 'л', 'l': 'д', ';': 'ж', '\'': 'э',
	'z': 'я', 'x': 'ч', 'c': 'с', 'v': 'м', 'b': 'и', 'n': 'т', 'm': 'ь', ',': 'б', '.': 'ю', '/': '.', '\\': '\\', '|': '/',
	'Q': 'Й', 'W': 'Ц', 'E': 'У', 'R': 'К', 'T': 'Е', 'Y': 'Н', 'U': 'Г', 'I': 'Ш', 'O': 'Щ', 'P': 'З', '{': 'Х', '}': 'Ъ',
	'A': 'Ф', 'S': 'Ы', 'D': 'В', 'F': 'А', 'G': 'П', 'H': 'Р', 'J': 'О', 'K': 'Л', 'L': 'Д', ':': 'Ж', '"': 'Э',
	'Z': 'Я', 'X': 'Ч', 'C': 'С', 'V': 'М', 'B': 'И', 'N': 'Т', 'M': 'Ь', '<': 'Б', '>': 'Ю', '?': ',',
}

var ruToEnMap = map[rune]rune{
	'ё': '`', 'Ё': '~', '"': '@', '№': '#', ';': '$', ':': '^', '?': '&',
	'й': 'q', 'ц': 'w', 'у': 'e', 'к': 'r', 'е': 't', 'н': 'y', 'г': 'u', 'ш': 'i', 'щ': 'o', 'з': 'p', 'х': '[', 'ъ': ']',
	'ф': 'a', 'ы': 's', 'в': 'd', 'а': 'f', 'п': 'g', 'р': 'h', 'о': 'j', 'л': 'k', 'д': 'l', 'ж': ';', 'э': '\'',
	'я': 'z', 'ч': 'x', 'с': 'c', 'м': 'v', 'и': 'b', 'т': 'n', 'ь': 'm', 'б': ',', 'ю': '.', '.': '/', '\\': '\\', '/': '|',
	'Й': 'Q', 'Ц': 'W', 'У': 'E', 'К': 'R', 'Е': 'T', 'Н': 'Y', 'Г': 'U', 'Ш': 'I', 'Щ': 'O', 'З': 'P', 'Х': '{', 'Ъ': '}',
	'Ф': 'A', 'Ы': 'S', 'В': 'D', 'А': 'F', 'П': 'G', 'Р': 'H', 'О': 'J', 'Л': 'K', 'Д': 'L', 'Ж': ':', 'Э': '"',
	'Я': 'Z', 'Ч': 'X', 'С': 'C', 'М': 'V', 'И': 'B', 'Т': 'N', 'Ь': 'M', 'Б': '<', 'Ю': '>', ',': '?',
}

func detectConversionDirection(r rune) conversionDirection {
	switch {
	case r == 'ё' || r == 'Ё' || (r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я'):
		return directionFromRussian
	case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
		return directionFromEnglish
	default:
		return directionUnknown
	}
}

func transformText(s string, defaultFromRussian bool) (string, bool) {
	defaultDir := directionFromEnglish
	if defaultFromRussian {
		defaultDir = directionFromRussian
	}
	currentDir := defaultDir
	lastExplicitDir := directionUnknown
	out := make([]rune, 0, len([]rune(s)))

	for _, r := range []rune(s) {
		explicitDir := detectConversionDirection(r)
		if explicitDir != directionUnknown {
			currentDir = explicitDir
			lastExplicitDir = explicitDir
		}

		var mapped rune
		var ok bool
		if currentDir == directionFromRussian {
			mapped, ok = ruToEnMap[r]
		} else {
			mapped, ok = enToRuMap[r]
		}
		if ok {
			out = append(out, mapped)
		} else {
			out = append(out, r)
		}

		switch r {
		case ' ', '\t', '\n', '\r':
			currentDir = defaultDir
		}
	}

	targetRussian := !defaultFromRussian
	switch lastExplicitDir {
	case directionFromRussian:
		targetRussian = false
	case directionFromEnglish:
		targetRussian = true
	}
	return string(out), targetRussian
}

func handleConvert() {
	if !atomic.CompareAndSwapInt32(&conversionBusy, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&conversionBusy, 0)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pOleInitialize.Call(0)
	defer pOleUninitialize.Call()

	normalizeModifiersBeforeCopy()

	fg, _, _ := pGetForegroundWindow.Call()
	if fg == 0 || fg == hwnd {
		return
	}
	tid, _, _ := pGetWindowThreadProcessId.Call(fg, 0)
	hkl, _, _ := pGetKeyboardLayout.Call(tid)
	langID := uint16(hkl & 0xffff)
	fromRussian := primaryLangID(langID) == LANG_RUSSIAN

	oldClipboard := saveClipboardObject()
	seqBefore, _, _ := pGetClipboardSequenceNumber.Call()
	sendCtrlKey(VK_C)

	changed := false
	for i := 0; i < 25; i++ {
		time.Sleep(8 * time.Millisecond)
		seqNow, _, _ := pGetClipboardSequenceNumber.Call()
		if seqNow != seqBefore {
			changed = true
			break
		}
	}
	if !changed {
		restoreClipboardObject(oldClipboard)
		return
	}

	selected, ok := getClipboardText()
	if !ok || selected == "" {
		restoreClipboardObject(oldClipboard)
		return
	}

	converted, targetRussian := transformText(selected, fromRussian)
	if !setClipboardText(converted) {
		restoreClipboardObject(oldClipboard)
		return
	}
	sendCtrlKey(VK_V)

	// Let the target control consume the paste before restoring the user's clipboard.
	time.Sleep(90 * time.Millisecond)
	restoreClipboardObject(oldClipboard)

	targetID := "00000409" // English (US)
	if targetRussian {
		targetID = "00000419" // Russian
	}
	targetHKL, _, _ := pLoadKeyboardLayoutW.Call(uintptr(unsafe.Pointer(utf16Ptr(targetID))), 0)
	if targetHKL != 0 {
		pPostMessageW.Call(fg, WM_INPUTLANGCHANGEREQUEST, 0, targetHKL)
	}
}

func startupKeyPath() *uint16   { return utf16Ptr(`Software\Microsoft\Windows\CurrentVersion\Run`) }
func startupValueName() *uint16 { return utf16Ptr("LayoutFixer") }

func startupEnabled() bool {
	var key uintptr
	r, _, _ := pRegOpenKeyExW.Call(HKEY_CURRENT_USER, uintptr(unsafe.Pointer(startupKeyPath())), 0, KEY_QUERY_VALUE, uintptr(unsafe.Pointer(&key)))
	if r != ERROR_SUCCESS {
		return false
	}
	defer pRegCloseKey.Call(key)
	var typ uint32
	var size uint32
	r, _, _ = pRegQueryValueExW.Call(key, uintptr(unsafe.Pointer(startupValueName())), 0, uintptr(unsafe.Pointer(&typ)), 0, uintptr(unsafe.Pointer(&size)))
	return r == ERROR_SUCCESS && typ == REG_SZ && size > 2
}

func setStartup(enable bool) {
	var key uintptr
	var disp uint32
	r, _, _ := pRegCreateKeyExW.Call(
		HKEY_CURRENT_USER,
		uintptr(unsafe.Pointer(startupKeyPath())),
		0, 0, 0,
		KEY_QUERY_VALUE|KEY_SET_VALUE,
		0,
		uintptr(unsafe.Pointer(&key)),
		uintptr(unsafe.Pointer(&disp)),
	)
	if r != ERROR_SUCCESS {
		return
	}
	defer pRegCloseKey.Call(key)

	if !enable {
		pRegDeleteValueW.Call(key, uintptr(unsafe.Pointer(startupValueName())))
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	value := `"` + exe + `"`
	u := utf16.Encode([]rune(value + "\x00"))
	pRegSetValueExW.Call(
		key,
		uintptr(unsafe.Pointer(startupValueName())),
		0,
		REG_SZ,
		uintptr(unsafe.Pointer(&u[0])),
		uintptr(len(u)*2),
	)
}

func showTrayMenu() {
	menu, _, _ := pCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer pDestroyMenu.Call(menu)

	pAppendMenuW.Call(menu, MF_STRING|MF_GRAYED, 0, uintptr(unsafe.Pointer(utf16Ptr("Layout Fixer — Win+Shift"))))
	pAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	flags := uintptr(MF_STRING)
	if startupEnabled() {
		flags |= MF_CHECKED
	}
	pAppendMenuW.Call(menu, flags, ID_STARTUP, uintptr(unsafe.Pointer(utf16Ptr("Run at Windows startup"))))
	pAppendMenuW.Call(menu, MF_SEPARATOR, 0, 0)
	pAppendMenuW.Call(menu, MF_STRING, ID_EXIT, uintptr(unsafe.Pointer(utf16Ptr("Exit"))))

	var pt POINT
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	pSetForegroundWindow.Call(hwnd)
	cmd, _, _ := pTrackPopupMenu.Call(menu, TPM_RIGHTBUTTON|TPM_RETURNCMD, uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	switch cmd {
	case ID_STARTUP:
		setStartup(!startupEnabled())
	case ID_EXIT:
		pDestroyWindow.Call(hwnd)
	}
}

func wndProc(h uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAY:
		if lParam == WM_RBUTTONUP || lParam == WM_LBUTTONUP {
			showTrayMenu()
			return 0
		}
	case WM_CONVERT:
		// Keep the hook/message-loop thread free while clipboard work happens.
		go handleConvert()
		return 0
	case WM_CLOSE:
		pDestroyWindow.Call(h)
		return 0
	case WM_DESTROY:
		removeTrayIcon()
		if hookHandle != 0 {
			pUnhookWindowsHookEx.Call(hookHandle)
		}
		pPostQuitMessage.Call(0)
		return 0
	case WM_COMMAND:
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(h, uintptr(msg), wParam, lParam)
	return r
}

func loadEmbeddedIcon(preferred int) uintptr {
	if len(embeddedIcon) < 22 || binary.LittleEndian.Uint16(embeddedIcon[2:4]) != 1 {
		return 0
	}
	count := int(binary.LittleEndian.Uint16(embeddedIcon[4:6]))
	best := -1
	bestDelta := 1 << 30
	for i := 0; i < count; i++ {
		off := 6 + i*16
		if off+16 > len(embeddedIcon) {
			break
		}
		w := int(embeddedIcon[off])
		h := int(embeddedIcon[off+1])
		if w == 0 {
			w = 256
		}
		if h == 0 {
			h = 256
		}
		d := w - preferred
		if d < 0 {
			d = -d
		}
		if h != w {
			d += 1000
		}
		if d < bestDelta {
			bestDelta = d
			best = off
		}
	}
	if best < 0 {
		return 0
	}
	size := int(binary.LittleEndian.Uint32(embeddedIcon[best+8 : best+12]))
	dataOff := int(binary.LittleEndian.Uint32(embeddedIcon[best+12 : best+16]))
	if size <= 0 || dataOff < 0 || dataOff+size > len(embeddedIcon) {
		return 0
	}
	ptr := uintptr(unsafe.Pointer(&embeddedIcon[dataOff]))
	h, _, _ := pCreateIconFromResourceEx.Call(
		ptr,
		uintptr(size),
		1,
		ICON_VERSION,
		uintptr(preferred),
		uintptr(preferred),
		LR_DEFAULTCOLOR,
	)
	return h
}

func addTrayIcon() bool {
	icon := appIcon
	if icon == 0 {
		icon = loadEmbeddedIcon(16)
	}
	if icon == 0 {
		icon = loadEmbeddedIcon(32)
	}
	if icon == 0 {
		icon, _, _ = pLoadIconW.Call(0, IDI_APPLICATION)
	}
	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hwnd
	nid.UID = ID_TRAY
	nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.UCallbackMessage = WM_TRAY
	nid.HIcon = icon
	copyUTF16(nid.SzTip[:], "Layout Fixer — Win+Shift")
	r, _, _ := pShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	return r != 0
}

func removeTrayIcon() {
	if hwnd == 0 {
		return
	}
	var nid NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hwnd
	nid.UID = ID_TRAY
	pShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}

func alreadyRunning() bool {
	name := utf16Ptr("Local\\LayoutFixer_EnglishRussian_6F1F01CB")
	h, _, _ := pCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if h == 0 {
		return false
	}
	errCode, _, _ := pGetLastError.Call()
	return errCode == ERROR_ALREADY_EXISTS
}

func main() {
	runtime.LockOSThread()
	if alreadyRunning() {
		pMessageBoxW.Call(0,
			uintptr(unsafe.Pointer(utf16Ptr("Layout Fixer is already running. Exit the old tray instance first, then start it again."))),
			uintptr(unsafe.Pointer(utf16Ptr("Layout Fixer"))),
			MB_OK|MB_ICONINFORMATION,
		)
		return
	}

	instance, _, _ := pGetModuleHandleW.Call(0)
	appIcon = loadEmbeddedIcon(16)
	if appIcon == 0 {
		appIcon = loadEmbeddedIcon(32)
	}
	if appIcon == 0 {
		appIcon, _, _ = pLoadIconW.Call(0, IDI_APPLICATION)
	}
	className := utf16Ptr("LayoutFixerHiddenWindow")
	wc := WNDCLASS{
		WndProc:   syscall.NewCallback(wndProc),
		Instance:  instance,
		Icon:      appIcon,
		ClassName: className,
	}
	atom, _, _ := pRegisterClassW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		return
	}

	hwnd, _, _ = pCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("Layout Fixer"))),
		0,
		0, 0, 0, 0,
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		return
	}

	if !addTrayIcon() {
		pDestroyWindow.Call(hwnd)
		return
	}

	hookProc := syscall.NewCallback(keyboardProc)
	hookHandle, _, _ = pSetWindowsHookExW.Call(WH_KEYBOARD_LL, hookProc, instance, 0)
	if hookHandle == 0 {
		pDestroyWindow.Call(hwnd)
		return
	}

	var msg MSG
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
	if appIcon != 0 {
		pDestroyIcon.Call(appIcon)
	}
}
