package main

import (
	"fmt"
	"net/http"
	"bufio"
	"log"
	"strconv"
	"strings"
	"sort"
	"math"
)

const NUM_GIBON_FILES int = 565
const NUM_SCOTT_FILES int = 393
const NUM_AUSTEN_FILES int = 414
const FILE_SUFFIX string = "One_"
const FILE_TYPE string = ".txx"
const SOURCE_URL string = "https://www.cs.brynmawr.edu/cs337/Lab04Data/"
const WORDS_START int = 3 // skip first 2 lines (contains chapter and chapter num)

type WordDocLoc struct {
	docId string // filename of file where word is found
	wordLoc int // line number where word is found
}

func (w WordDocLoc) String() string {
	return fmt.Sprintf("{%s %d}", w.docId, w.wordLoc)
}

/* reads the file that corresponds to the given file info
 * and returns the document id and a string slice of lines
 */
func readFile(author string, fileNum int) (string, []string) {
	docId := author + FILE_SUFFIX + strconv.Itoa(fileNum) + FILE_TYPE
	resp, httpErr := http.Get(SOURCE_URL + docId)

	if httpErr != nil {
		log.Fatal(httpErr)
	}

	defer resp.Body.Close()

	var lines []string
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	scannerErr := scanner.Err()
	if scannerErr != nil {
		log.Fatal(scannerErr)
	}

	return docId, lines
}

/* given that each line contains a single word, parse the given slice of lines
 * and populate a map of each word to its list of wordDocLocs (file and line num)
*/
func parseFile(docId string, lines []string, wordMap map[string][]WordDocLoc) {
	wordLoc := WORDS_START
	for i := 2; i < len(lines); i++ {
		word := strings.ToLower(lines[i])
		val, valExists := wordMap[word];
		if !valExists {
			wordMap[word] = make([]WordDocLoc, 5) // initialize a new slice w/ default size 5
		}
		// add word's wordDocLoc to its value slice
		wordMap[word] = append(val, WordDocLoc{docId: docId, wordLoc: wordLoc})
		wordLoc += 1
	}
}

// find and return a list of files that contain the entire set of given words
func findFileContaining(words []string, wordMap map[string][]WordDocLoc) []string {
	// initialize a map of docIds to how many words from words slice they contain
	docToWordCount := make(map[string]int)
	for _, word := range words { // for each of the given set of words
		wordDocLocs := wordMap[word] // get its list wordDocLocs
		// find the unique doc ids in the wordDocLocs
		uniqueDocIds := make(map[string]bool)
		for _, wordDocLoc := range wordDocLocs {
			docId := wordDocLoc.docId
			if _, valExists := uniqueDocIds[docId]; !valExists {
				uniqueDocIds[docId] = true
			}
		}

		// for each unique doc id found, increment its count in the map
		for uniqueDocId, _ := range uniqueDocIds {
			if _, valExists := docToWordCount[uniqueDocId]; !valExists {
				docToWordCount[uniqueDocId] = 0
			}
			docToWordCount[uniqueDocId]++
		}
	}

	var files []string // initialize list of files that contain the entire set of words
	for docId, wordCount := range docToWordCount { // for each docId in the map
		if wordCount == len(words) { // if a docId is associated with as many words as in the set
			files = append(files, docId) // add to list of files to return
		}
	}

	return files
}

/* reorders the given list of files by ranking,
 * where "better" files are put first (a "better" file can be defined as a file in which the
 * set of words it is meant to contain are located closer together; for now assuming 2 words)
 * returns a map of file to the distance between the two given words; map is not sorted
 */
func rankFilesContaining(files []string, word1 string, word2 string, wordMap map[string][]WordDocLoc) map[string]float64 {
	fileToDistance := make(map[string]float64)

	sort.Slice(files, func(x int, y int) bool {
		smallestDistance1 := findSmallestDistanceBetween(word1, word2, files[x], wordMap)
		smallestDistance2 := findSmallestDistanceBetween(word1, word2, files[y], wordMap)
		fileToDistance[files[x]] = smallestDistance1
		fileToDistance[files[y]] = smallestDistance2
		return smallestDistance1 < smallestDistance2
	})

	return fileToDistance
}

// finds the smallest distance between 2 given words in a given file
func findSmallestDistanceBetween(word1 string, word2 string, file string, wordMap map[string][]WordDocLoc) float64 {
	wordDocLocs1, _ := wordMap[word1]
	var word1LocsInFile []int // list of locations of word1 in the given file
	for _, wordDocLoc := range wordDocLocs1 {
		if wordDocLoc.docId == file {
			word1LocsInFile = append(word1LocsInFile, wordDocLoc.wordLoc)
		}
	}
	sort.Slice(word1LocsInFile, func(x int, y int) bool {
		return word1LocsInFile[x] < word1LocsInFile[y]
	})

	wordDocLocs2, _ := wordMap[word2]
	var word2LocsInFile []int // list of locations of word2 in the given file
	for _, wordDocLoc := range wordDocLocs2 {
		if wordDocLoc.docId == file {
			word2LocsInFile = append(word2LocsInFile, wordDocLoc.wordLoc)
		}
	}
	sort.Slice(word2LocsInFile, func(x int, y int) bool {
		return word2LocsInFile[x] < word2LocsInFile[y]
	})

	i1 := 0 // index in word1LocsInFile
	i2 := 0 // index in word2LocsInFile
	smallestDistance := math.Abs(float64(word1LocsInFile[i1]) - float64(word2LocsInFile[i2])) // distance between the first items

	for i1 < len(word1LocsInFile) && i2 < len(word2LocsInFile) {
		distance := math.Abs(float64(word1LocsInFile[i1]) - float64(word2LocsInFile[i2]))
		if distance < smallestDistance {
			smallestDistance = distance
		}

		if word1LocsInFile[i1] < word2LocsInFile[i2] {
			i1 += 1
		} else {
			i2 += 1
		}
	}

	return smallestDistance
}

// prints a map of ranked files given list of files in correct order
func printRankedFiles(files []string, rankedFiles map[string]float64) {
	for i, file := range files {
		fmt.Printf("%d %6d %s\n", i+1, int(rankedFiles[file]), file)
	}
}

func main() {
	wordMap := make(map[string][]WordDocLoc)

	// read and parse all gibbon files
	for i := 1; i <= NUM_GIBON_FILES; i++ {
		docId, lines := readFile("Gibon", i)
		parseFile(docId, lines, wordMap)
	}
	
	// read and parse all scott files
	for i := 1; i <= NUM_SCOTT_FILES; i++ {
		docId, lines := readFile("Scott", i)
		parseFile(docId, lines, wordMap)
	}

	// read and parse all austen files
	for i := 1; i <= NUM_AUSTEN_FILES; i++ {
		docId, lines := readFile("Austen", i)
		parseFile(docId, lines, wordMap)
	}

	// test searching files for pairs gets correct files, correctly ranked by distance
	pair1 := []string{"elizabeth", "emma"}
	pair2 := []string{"roy", "clan"}
	pair3 := []string{"legend", "legion"}

	files1 := findFileContaining(pair1, wordMap)
	files2 := findFileContaining(pair2, wordMap)
	files3 := findFileContaining(pair3, wordMap)

	fmt.Println(pair1)
	printRankedFiles(files1, rankFilesContaining(files1, pair1[0], pair1[1], wordMap))
	fmt.Printf("\n%v\n", pair2)
	printRankedFiles(files2, rankFilesContaining(files2, pair2[0], pair2[1], wordMap))
	fmt.Printf("\n%v\n", pair3)
	printRankedFiles(files3, rankFilesContaining(files3, pair3[0], pair3[1], wordMap))
}
