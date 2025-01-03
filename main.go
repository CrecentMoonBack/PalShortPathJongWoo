package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ----------------------------------------------------
// 1) 데이터 구조 정의
// ----------------------------------------------------

// Comb: 한 줄의 교배식 (부모1, 부모2, 자식)
type Comb struct {
	Parents [2]string
	Child   string
}

// Breed: 한 Node(팰)가 다른 팰(spouse)과 교배해서 낳은 자식(child)
type Breed struct {
	Spouse string
	Child  string
}

// Node: 팰(몬스터). 이름(name)과 교배 통해 낳은 자식 정보(breeds)
type Node struct {
	Name   string
	Breeds []Breed
}

// Link: BFS 탐색 시 각 단계(노드)에서의 상태 저장
type Link struct {
	NodeIdx      int
	PrevLinkIdx  int
	PrevChildIdx int
	LinkNum      int
}

// queue: BFS용 간단 큐
type queue struct {
	data []int
}

func (q *queue) push(x int) {
	q.data = append(q.data, x)
}

func (q *queue) pop() (int, bool) {
	if len(q.data) == 0 {
		return 0, false
	}
	x := q.data[0]
	q.data = q.data[1:]
	return x, true
}
func (q *queue) empty() bool {
	return len(q.data) == 0
}

// ----------------------------------------------------
// 2) 파일 읽기: readFile
// ----------------------------------------------------

func readFile(path string) ([]Comb, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("입력 파일(%s)을 열 수 없습니다: %w", path, err)
	}
	defer file.Close()

	var combs []Comb
	var names []string              // 이름들
	nameMap := make(map[string]int) // 이름 -> 인덱스

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		replaced := strings.ReplaceAll(line, "+", " ")
		replaced = strings.ReplaceAll(replaced, "=", " ")
		parts := strings.Fields(replaced) // 공백 분할

		if len(parts) != 3 {
			// 잘못된 형식이면 무시
			continue
		}

		p1 := strings.TrimSpace(parts[0])
		p2 := strings.TrimSpace(parts[1])
		c := strings.TrimSpace(parts[2])

		if p1 == "" || p2 == "" || c == "" {
			// 이상한 데이터
			continue
		}

		if _, ok := nameMap[p1]; !ok {
			nameMap[p1] = len(names)
			names = append(names, p1)
		}
		if _, ok := nameMap[p2]; !ok {
			nameMap[p2] = len(names)
			names = append(names, p2)
		}
		if _, ok := nameMap[c]; !ok {
			nameMap[c] = len(names)
			names = append(names, c)
		}

		combs = append(combs, Comb{
			Parents: [2]string{p1, p2},
			Child:   c,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	// 중복/충돌 검사
	duplicateIndices := make(map[int]bool)
	n := len(combs)
	for i := 0; i < n; i++ {
		if duplicateIndices[i] {
			continue
		}
		for j := i + 1; j < n; j++ {
			if duplicateIndices[j] {
				continue
			}
			comb1 := combs[i]
			comb2 := combs[j]

			isSameParents := false
			if (comb1.Parents[0] == comb2.Parents[0] && comb1.Parents[1] == comb2.Parents[1]) ||
				(comb1.Parents[0] == comb2.Parents[1] && comb1.Parents[1] == comb2.Parents[0]) {
				isSameParents = true
			}
			if !isSameParents {
				continue
			}
			// 부모가 같은데 자식 다르면 충돌
			if comb1.Child != comb2.Child {
				return nil, nil, errors.New("[교배식 오류] 같은 부모 조합에 서로 다른 자식이 있습니다.")
			} else {
				// 중복
				duplicateIndices[j] = true
			}
		}
	}

	// 중복 제거
	filtered := make([]Comb, 0, n)
	for i, c := range combs {
		if !duplicateIndices[i] {
			filtered = append(filtered, c)
		}
	}
	return filtered, names, nil
}

// ----------------------------------------------------
// 3) Node 구성: constructNodes
// ----------------------------------------------------
func constructNodes(combs []Comb) ([]Node, error) {
	nodeIndex := make(map[string]int)
	var nodes []Node

	for _, comb := range combs {
		p0 := comb.Parents[0]
		p1 := comb.Parents[1]
		c := comb.Child

		isSameParent := (p0 == p1)
		parentNum := 2
		if isSameParent {
			parentNum = 1
		}

		for i := 0; i < parentNum; i++ {
			var parentName, spouseName string
			if i == 0 {
				parentName = p0
				spouseName = p1
			} else {
				parentName = p1
				spouseName = p0
			}

			ndx, ok := nodeIndex[parentName]
			if !ok {
				ndx = len(nodes)
				nodeIndex[parentName] = ndx
				nodes = append(nodes, Node{
					Name:   parentName,
					Breeds: []Breed{},
				})
			}

			nodes[ndx].Breeds = append(nodes[ndx].Breeds, Breed{
				Spouse: spouseName,
				Child:  c,
			})
		}
	}
	return nodes, nil
}

// ----------------------------------------------------
// 4) BFS: searchShortestPaths
// ----------------------------------------------------

func findNodeIdx(nodes []Node, name string) int {
	for i := range nodes {
		if nodes[i].Name == name {
			return i
		}
	}
	return -1
}

func searchShortestPaths(searchParent, searchChild string, maxLinkNum int, nodes []Node) ([]int, []Link, error) {
	var linkBuffer []Link
	var Q queue

	// 시작 노드(부모) 찾아 큐에 삽입
	for i := range nodes {
		if nodes[i].Name == searchParent {
			linkBuffer = append(linkBuffer, Link{
				NodeIdx:      i,
				PrevLinkIdx:  -1,
				PrevChildIdx: -1,
				LinkNum:      0,
			})
			Q.push(len(linkBuffer) - 1)
		}
	}

	isFound := false
	foundLinkNum := -1
	var foundIndices []int

	for !Q.empty() {
		linkIdx, _ := Q.pop()
		link := linkBuffer[linkIdx]

		if isFound && link.LinkNum > foundLinkNum {
			continue
		}
		if link.LinkNum >= maxLinkNum {
			continue
		}

		// linkNum != 0 이면 도착 체크
		if link.LinkNum != 0 {
			if nodes[link.NodeIdx].Name == searchChild {
				if !isFound {
					isFound = true
					foundLinkNum = link.LinkNum
				}
				foundIndices = append(foundIndices, linkIdx)
				continue
			}
		}

		// 자식으로 이동
		node := nodes[link.NodeIdx]
		for childIdx, b := range node.Breeds {
			nxt := findNodeIdx(nodes, b.Child)
			if nxt == -1 {
				continue
			}
			newLink := Link{
				NodeIdx:      nxt,
				PrevLinkIdx:  linkIdx,
				PrevChildIdx: childIdx,
				LinkNum:      link.LinkNum + 1,
			}
			linkBuffer = append(linkBuffer, newLink)
			Q.push(len(linkBuffer) - 1)
		}
	}

	return foundIndices, linkBuffer, nil
}

// ----------------------------------------------------
// 5) 경로 복원: makePathString
// ----------------------------------------------------

func makePathString(nodes []Node, linkBuffer []Link, foundLinkIdx int) string {
	var path []Link
	cur := foundLinkIdx
	for cur != -1 {
		path = append(path, linkBuffer[cur])
		cur = linkBuffer[cur].PrevLinkIdx
	}

	// path[0] = 마지막 지점, path[len-1] = 시작점
	var sb strings.Builder
	step := 1
	// 맨 뒤가 시작점 → linkNum=0
	// 따라서 len(path)-2 부터 0까지
	for i := len(path) - 2; i >= 0; i-- {
		link := path[i]
		prevLink := path[i+1]
		node := nodes[link.NodeIdx]
		prevNode := nodes[prevLink.NodeIdx]

		spouse := prevNode.Breeds[link.PrevChildIdx].Spouse
		child := node.Name
		parent1 := prevNode.Name
		parent2 := spouse

		sb.WriteString(fmt.Sprintf("%d) \"%s\" + \"%s\" = \"%s\"\n", step, parent1, parent2, child))
		step++
	}
	return sb.String()
}

// ----------------------------------------------------
// 6) main 함수 (Fyne GUI)
// ----------------------------------------------------
var combs []Comb
var allNodes []Node

func main() {
	// 콘솔 숨기려면, 빌드 시 -ldflags "-H=windowsgui" 옵션을 사용
	// 아이콘 포함 위해 .syso 리소스 파일 준비 (winres or rsrc 등)

	myApp := app.New()
	myWindow := myApp.NewWindow("교배식 최단 경로 찾기")
	icon, err := fyne.LoadResourceFromPath("123.ico")
	if err != nil {
		log.Println("아이콘 로드 실패:", err)
	} else {
		// 앱 전체 아이콘 지정 (macOS 등에서 Dock 아이콘으로 쓰일 수 있음)
		myApp.SetIcon(icon)
		// 윈도우 타이틀바 아이콘 지정 (Windows 등)
		myWindow.SetIcon(icon)
	}

	myWindow.Resize(fyne.NewSize(700, 500))

	labelInfo := widget.NewLabel("안내: \"교배식.txt\" 파일을 기반으로, 부모 팰 → 자식 팰 최단 교배 경로를 찾아줍니다.\n" +
		"교배식 파일을 연 뒤, 부모/자식을 입력해 검색하거나, 교배식을 추가할 수 있습니다.")

	entryParent := widget.NewEntry()
	entryParent.SetPlaceHolder("부모 팰 입력")
	entryChild := widget.NewEntry()
	entryChild.SetPlaceHolder("자식 팰 입력")

	resultArea := widget.NewMultiLineEntry()
	resultArea.SetMinRowsVisible(10)
	// 만약 Fyne 버전이 v2.4 이상이라면:
	// resultArea.SetReadOnly(true)

	// 전역 상태
	var openedFilePath string // 사용자가 연 파일 경로

	// 1) 파일 열기 버튼
	openBtn := widget.NewButton("교배식 파일 열기", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if r == nil {
				return
			}
			filePath := r.URI().Path()
			c, _, e := readFile(filePath)
			if e != nil {
				dialog.ShowError(e, myWindow)
				return
			}
			ns, e2 := constructNodes(c)
			if e2 != nil {
				dialog.ShowError(e2, myWindow)
				return
			}

			combs = c
			allNodes = ns
			openedFilePath = filePath

			dialog.ShowInformation("알림", "파일을 성공적으로 불러왔습니다.", myWindow)
		}, myWindow)
	})

	// 2) 검색 버튼
	searchBtn := widget.NewButton("검색", func() {
		if openedFilePath == "" {
			dialog.ShowInformation("안내", "먼저 교배식 파일을 열어주세요.", myWindow)
			return
		}

		// 파일 다시 로드
		newCombs, _, err := readFile(openedFilePath)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		newNodes, err2 := constructNodes(newCombs)
		if err2 != nil {
			dialog.ShowError(err2, myWindow)
			return
		}

		// 검색하기 전에 사용자 입력 확인
		parentName := strings.TrimSpace(entryParent.Text)
		childName := strings.TrimSpace(entryChild.Text)
		if parentName == "" || childName == "" {
			dialog.ShowInformation("안내", "부모 팰과 자식 팰 이름을 모두 입력해주세요.", myWindow)
			return
		}

		if findNodeIdx(newNodes, parentName) == -1 || findNodeIdx(newNodes, childName) == -1 {
			dialog.ShowInformation("안내", "교배식에 없는 팰 이름입니다.", myWindow)
			return
		}

		foundIndices, linkBuf, err3 := searchShortestPaths(parentName, childName, 10, newNodes)
		if err3 != nil {
			dialog.ShowError(err3, myWindow)
			return
		}

		if len(foundIndices) == 0 {
			resultArea.SetText("찾을 수 없습니다. (교배 경로가 존재하지 않음)")
			return
		}

		var sb strings.Builder
		for i, idx := range foundIndices {
			sb.WriteString(fmt.Sprintf("교배식 경로 %d\n", i+1))
			pathStr := makePathString(newNodes, linkBuf, idx)
			sb.WriteString(pathStr + "\n")
		}
		resultArea.SetText(sb.String())
	})

	// 3) 교배식 추가 버튼
	addBtn := widget.NewButton("교배식 추가", func() {
		if openedFilePath == "" {
			dialog.ShowInformation("안내", "먼저 교배식 파일을 열어주세요.", myWindow)
			return
		}

		// 새 교배식 입력받는 팝업
		var entryP1, entryP2, entryC *widget.Entry
		entryP1 = widget.NewEntry()
		entryP2 = widget.NewEntry()
		entryC = widget.NewEntry()

		content := container.NewVBox(
			widget.NewLabel("새로운 교배식을 추가합니다.\n(부모1 + 부모2 = 자식)"),
			entryP1, entryP2, entryC,
		)

		dialog.ShowCustomConfirm("교배식 추가", "추가", "취소", content, func(confirm bool) {
			if confirm {
				p1 := strings.TrimSpace(entryP1.Text)
				p2 := strings.TrimSpace(entryP2.Text)
				ch := strings.TrimSpace(entryC.Text)

				// 간단 유효성 검사
				if p1 == "" || p2 == "" || ch == "" {
					dialog.ShowInformation("안내", "모든 항목을 입력해주세요.", myWindow)
					return
				}

				// 파일에 추가
				f, err := os.OpenFile(openedFilePath, os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					dialog.ShowError(err, myWindow)
					return
				}
				defer f.Close()

				lineToAdd := fmt.Sprintf("\n%s+%s=%s", p1, p2, ch)
				_, err = f.WriteString(lineToAdd)
				if err != nil {
					dialog.ShowError(err, myWindow)
					return
				}

				// 전체 재로딩
				newCombs, _, e := readFile(openedFilePath)
				if e != nil {
					dialog.ShowError(e, myWindow)
					return
				}
				newNodes, e2 := constructNodes(newCombs)
				if e2 != nil {
					dialog.ShowError(e2, myWindow)
					return
				}

				combs = newCombs
				allNodes = newNodes

				dialog.ShowInformation("알림", "교배식을 성공적으로 추가했습니다!", myWindow)
			}
		}, myWindow)
	})

	content := container.NewVBox(
		labelInfo,
		container.NewGridWithColumns(2,
			entryParent,
			entryChild,
		),
		container.NewHBox(
			openBtn,
			searchBtn,
			addBtn,
		),
		resultArea,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
