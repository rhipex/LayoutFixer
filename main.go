//go:build windows

package main

import (
	_ "embed"
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

//go:embed layoutfixer.ico
var embeddedIcon []byte

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

func transformText(s string, fromRussian bool) string {
	enToRu := map[rune]rune{
		'`': 'ё', '~': 'Ё', '@': '"', '#': '№', '$': ';', '^': ':', '&': '?',
		'q': 'й', 'w': 'ц', 'e': 'у', 'r': 'к', 't': 'е', 'y': 'н', 'u': 'г', 'i': 'ш', 'o': 'щ', 'p': 'з', '[': 'х', ']': 'ъ',
		'a': 'ф', 's': 'ы', 'd': 'в', 'f': 'а', 'g': 'п', 'h': 'р', 'j': 'о', 'k': 'л', 'l': 'д', ';': 'ж', '\'': 'э',
		'z': 'я', 'x': 'ч', 'c': 'с', 'v': 'м', 'b': 'и', 'n': 'т', 'm': 'ь', ',': 'б', '.': 'ю', '/': '.', '\\': '\\', '|': '/',
		'Q': 'Й', 'W': 'Ц', 'E': 'У', 'R': 'К', 'T': 'Е', 'Y': 'Н', 'U': 'Г', 'I': 'Ш', 'O': 'Щ', 'P': 'З', '{': 'Х', '}': 'Ъ',
		'A': 'Ф', 'S': 'Ы', 'D': 'В', 'F': 'А', 'G': 'П', 'H': 'Р', 'J': 'О', 'K': 'Л', 'L': 'Д', ':': 'Ж', '"': 'Э',
		'Z': 'Я', 'X': 'Ч', 'C': 'С', 'V': 'М', 'B': 'И', 'N': 'Т', 'M': 'Ь', '<': 'Б', '>': 'Ю', '?': ',',
	}
	ruToEn := map[rune]rune{
		'ё': '`', 'Ё': '~', '"': '@', '№': '#', ';': '$', ':': '^', '?': '&',
		'й': 'q', 'ц': 'w', 'у': 'e', 'к': 'r', 'е': 't', 'н': 'y', 'г': 'u', 'ш': 'i', 'щ': 'o', 'з': 'p', 'х': '[', 'ъ': ']',
		'ф': 'a', 'ы': 's', 'в': 'd', 'а': 'f', 'п': 'g', 'р': 'h', 'о': 'j', 'л': 'k', 'д': 'l', 'ж': ';', 'э': '\'',
		'я': 'z', 'ч': 'x', 'с': 'c', 'м': 'v', 'и': 'b', 'т': 'n', 'ь': 'm', 'б': ',', 'ю': '.', '.': '/', '\\': '\\', '/': '|',
		'Й': 'Q', 'Ц': 'W', 'У': 'E', 'К': 'R', 'Е': 'T', 'Н': 'Y', 'Г': 'U', 'Ш': 'I', 'Щ': 'O', 'З': 'P', 'Х': '{', 'Ъ': '}',
		'Ф': 'A', 'Ы': 'S', 'В': 'D', 'А': 'F', 'П': 'G', 'Р': 'H', 'О': 'J', 'Л': 'K', 'Д': 'L', 'Ж': ':', 'Э': '"',
		'Я': 'Z', 'Ч': 'X', 'С': 'C', 'М': 'V', 'И': 'B', 'Т': 'N', 'Ь': 'M', 'Б': '<', 'Ю': '>', ',': '?',
	}
	out := make([]rune, 0, len([]rune(s)))
	for _, r := range []rune(s) {
		var mapped rune
		var ok bool
		if fromRussian {
			mapped, ok = ruToEn[r]
		} else {
			mapped, ok = enToRu[r]
		}
		if ok {
			out = append(out, mapped)
		} else {
			out = append(out, r)
		}
	}
	return string(out)
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

	converted := transformText(selected, fromRussian)
	if !setClipboardText(converted) {
		restoreClipboardObject(oldClipboard)
		return
	}
	sendCtrlKey(VK_V)

	// Let the target control consume the paste before restoring the user's clipboard.
	time.Sleep(90 * time.Millisecond)
	restoreClipboardObject(oldClipboard)

	targetID := "00000419" // Russian
	if fromRussian {
		targetID = "00000409" // English (US)
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

	pAppendMenuW.Call(menu, MF_STRING|MF_GRAYED, 0, uintptr(unsafe.Pointer(utf16Ptr("Layout Fixer v3 — Win+Shift"))))
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
	copyUTF16(nid.SzTip[:], "Layout Fixer v3 — Win+Shift")
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
			uintptr(unsafe.Pointer(utf16Ptr("Layout Fixer is already running. Exit the old tray instance first, then start LayoutFixer v3."))),
			uintptr(unsafe.Pointer(utf16Ptr("Layout Fixer v3"))),
			MB_OK|MB_ICONINFORMATION,
		)
		return
	}

	instance, _, _ := pGetModuleHandleW.Call(0)
	// Resource ID 1 is the RT_GROUP_ICON embedded into the PE at build time.
	appIcon, _, _ = pLoadIconW.Call(instance, 1)
	if appIcon == 0 {
		appIcon = loadEmbeddedIcon(32)
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
