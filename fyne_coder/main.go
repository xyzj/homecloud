package main

import (
	// static
	lib "coder/lib"
	_ "embed"

	fynex "coder/fyne"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//go:embed favicon.png
var favicon []byte

var topWindow fyne.Window

func main() {
	a := fynex.NewDesktopApp(&fynex.DesktopAppOptions{
		AppID:       "tools.coder.xy",
		WindowTitle: "字符串编码工具",
		Icon:        fyne.NewStaticResource("favicon.png", favicon),
		Size:        fyne.NewSize(1080, 800),
	})
	topWindow = a.MainWindow()

	lib.InitCards()
	card := widget.NewCard("Welcome", "一堆字符串工具", nil)
	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return lib.NavIdx[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := lib.NavIdx[uid]
			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("unknow name")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := lib.NavItems[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
			if unsupportedTutorial(t) {
				obj.(*widget.Label).TextStyle = fyne.TextStyle{Italic: true}
			} else {
				obj.(*widget.Label).TextStyle = fyne.TextStyle{}
			}
		},
		OnSelected: func(uid widget.TreeNodeID) {
			if t, ok := lib.NavItems[uid]; ok {
				if unsupportedTutorial(t) {
					return
				}

				if fyne.CurrentDevice().IsMobile() {
					child := a.MainApp().NewWindow(t.Title)
					topWindow = child
					child.SetContent(t.View(topWindow))
					child.Show()
					child.SetOnClosed(func() {
						topWindow = a.MainWindow()
					})
					return
				}
				card.SetTitle(t.Title)
				card.SetSubTitle(t.Intro)
				if t.View != nil {
					card.SetContent(t.View(nil))
				} else {
					if t.Object != nil {
						card.SetContent(t.Object)
					}
				}
				card.Refresh()
			}
		},
	}
	tree.OpenAllBranches()
	var cont fyne.CanvasObject
	if fyne.CurrentDevice().IsMobile() {
		cont = container.NewStack(tree)
	} else {
		dcont := container.NewHSplit(tree, card)
		dcont.SetOffset(0.2)
		cont = dcont
	}
	a.ShowAndRun(cont)
}
func unsupportedTutorial(t *lib.NavItem) bool {
	return !t.SupportWeb && fyne.CurrentDevice().IsBrowser()
}
