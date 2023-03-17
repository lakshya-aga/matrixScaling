package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
)

func createList(data [][]string) [][]float64 {
	var elements [][]float64
	for _, line := range data {
		// omit header line
		var rec []float64
		for _, field := range line {
			if field == "" {
				rec = append(rec, math.NaN())
				continue
			}
			n, _ := strconv.ParseFloat(field, 64)
			//print(n)
			rec = append(rec, n)
		}

		elements = append(elements, rec)
	}
	return elements
}
func getMean(data []float64) float64 {
	sum := 0.0
	count := 0
	for _, j := range data {
		if j != math.NaN() {
			count++
			sum = sum + j
		}
	}
	return sum / float64(count)
}

func getMean2(data [][]float64) float64 {
	sum := 0.0
	count := 0
	for i, j := range data {
		for k, _ := range j {
			if math.IsNaN(data[i][k]) {
				continue
			}

			count++
			sum = sum + data[i][k]

		}
	}
	return sum / float64(count)
}
func getSD2(data [][]float64) float64 {
	mean := getMean2(data)
	sum := 0.0
	count := 0
	for _, j := range data {
		for _, l := range j {
			if math.IsNaN(l) {
				continue
			}
			count++
			sum = sum + (l-mean)*(l-mean)

		}
	}
	return math.Sqrt(sum / float64(count))
}
func scaleList(data [][]float64) [][]float64 {
	var a []float64
	var ans [][]float64
	mean := getMean2(data)
	sd := getSD2(data)

	for i, l1 := range data {
		a = []float64{}
		for j, _ := range l1 {
			if data[i][j] != math.NaN() {
				a = append(a, ((data[i][j] - mean) / sd))
			}
		}
		ans = append(ans, a)
	}
	return ans
}

// Create the change in Alpha array for every column
func getAlphaChange(scaled_list [][]float64, tao []float64, gamma []float64) []float64 {
	var ans []float64

	for i, row := range scaled_list {
		sum1 := 0.0
		sum2 := 0.0
		count := 0
		for j, item := range row {
			if math.IsNaN(item) {
				continue
			}
			sum1 = sum1 + item
			sum2 = sum2 + 1/tao[i]*gamma[j]
			count++
		}
		ans = append(ans, sum1/sum2)
	}
	return ans
}

// Create the change in beta array for every column
func getBetaChange(scaled_list [][]float64, tao []float64, gamma []float64) []float64 {
	var ans []float64
	for i, _ := range scaled_list[0] {
		sum1 := 0.0
		sum2 := 0.0
		count := 0
		for j, _ := range scaled_list {
			if math.IsNaN(scaled_list[j][i]) {
				continue
			}
			//no need to check for tao and gamma to be NaN since a single numeric entry ensures the row and column have all parameters not equal to NaN
			sum1 = sum1 + scaled_list[j][i]
			sum2 = sum2 + 1/tao[i]*gamma[j]
			count++
		}
		ans = append(ans, sum1/sum2)
	}
	return ans
}

func getTaoChange(scaled_list [][]float64) []float64 {
	var ans []float64
	for i, _ := range scaled_list {
		sum1 := 0.0
		count := 0
		for j, _ := range scaled_list[0] {
			if math.IsNaN(scaled_list[i][j]) {
				continue
			}
			//no need to check for tao and gamma to be NaN since a single numeric entry ensures the row and column have all parameters not equal to NaN
			sum1 = sum1 + scaled_list[i][j]*scaled_list[i][j]
			count++
		}
		ans = append(ans, math.Sqrt(sum1/float64(count)))
	}
	return ans
}

func getGammaChange(scaled_list [][]float64) []float64 {
	var ans []float64
	for i, _ := range scaled_list[0] {
		sum1 := 0.0
		count := 0
		for j, _ := range scaled_list {
			if math.IsNaN(scaled_list[j][i]) {
				continue
			}
			//no need to check for tao and gamma to be NaN since a single numeric entry ensures the row and column have all parameters not equal to NaN
			sum1 = sum1 + scaled_list[j][i]*scaled_list[j][i]
			count++
		}
		ans = append(ans, math.Sqrt(sum1/float64(count)))
	}
	return ans
}

func rescale(matrix [][]float64, alpha []float64, beta []float64, gamma []float64, tao []float64) [][]float64 {
	var ans [][]float64
	for i, row := range matrix {
		var a []float64
		for j, item := range row {
			a = append(a, (item-alpha[i]-beta[j])/gamma[j]/tao[i])
		}
		ans = append(ans, a)
	}
	return ans
}
func calculateHeuristic(alpha []float64, beta []float64, gamma []float64, tao []float64) float64 {
	sum1 := 0.0
	sum2 := 0.0
	sum3 := 0.0
	sum4 := 0.0
	for _, j := range alpha {
		sum1 = sum1 + j*j
	}
	for _, j := range beta {
		sum2 = sum2 + j*j
	}
	for _, j := range gamma {
		sum3 = sum3 + math.Log(j)*math.Log(j)
	}
	for _, j := range tao {
		sum4 = sum4 + math.Log(j)*math.Log(j)
	}

	return (sum1 + sum2 + sum3 + sum4)
}
func main() {
	// open file
	flag.Usage = func() {
		fmt.Printf("Usage: %s [option1] <csvFile> [option2] <csvFile>\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	if len(os.Args) < 2 {
		println("A filepath argument is required")
		return
	}
	in := flag.String("i", "data.csv", "Input file path")
	o := flag.String("o", "convergence.csv", "output file path")
	flag.Parse()
	f, err := os.Open(*in)
	//Logging error if any in opening the input csv file
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// convert records to array of structs
	input_list := createList(data)
	//fmt.Printf("%+v\n", input_list)
	scaled_list := scaleList(input_list)
	//println()
	//fmt.Printf("%+v\n", scaled_list)

	sum := 0.0
	//Row means in alpha
	var alpha []float64
	count := 0
	for _, l := range scaled_list {
		count = 0
		sum = 0.0
		for _, l2 := range l {

			if math.IsNaN(l2) {
				continue
			}
			count = count + 1
			sum = sum + l2

		}
		alpha = append(alpha, sum/float64(count))

	}
	//Column means in beta
	var beta []float64
	count = 0
	sum = 0.0

	for i, _ := range scaled_list[0] {
		count = 0
		sum = 0.0
		for j, _ := range scaled_list {
			if math.IsNaN(scaled_list[j][i]) {
				continue
			}
			sum = sum + scaled_list[j][i]
			count = count + 1

		}
		beta = append(beta, sum/float64(count))
	}
	//Row standard deviations
	var tao []float64
	for j, row := range scaled_list {
		count = 0
		sum = 0.0
		for _, elem := range row {
			if math.IsNaN(elem) {
				continue
			}
			count = count + 1
			sum = sum + (elem-alpha[j])*(elem-alpha[j])
		}
		tao = append(tao, math.Sqrt(sum/float64(count)))
	}

	//Column Standard Deviations
	var gamma []float64
	for i, _ := range scaled_list[0] {
		count = 0
		sum = 0.0
		for j, _ := range scaled_list {
			if math.IsNaN(scaled_list[j][i]) {
				continue
			}
			count = count + 1
			sum = sum + (scaled_list[j][i]-beta[i])*(scaled_list[j][i]-beta[i])

		}
		gamma = append(gamma, math.Sqrt(sum/float64(count)))
	}
	var delta_alpha []float64
	var delta_beta []float64
	var delta_gamma []float64
	var delta_tao []float64
	heuristic1 := calculateHeuristic(delta_alpha, delta_beta, delta_gamma, delta_tao)
	heuristic2 := 0.0
	for {
		//println(heuristic1)
		delta_alpha = getAlphaChange(scaled_list, tao, gamma)
		//fmt.Println(delta_alpha)
		//For debugging

		delta_beta = getBetaChange(scaled_list, tao, gamma)
		//fmt.Println(delta_beta)
		//For debugging

		delta_gamma = getGammaChange(scaled_list)
		//fmt.Println(delta_gamma)
		//For debugging

		delta_tao = getTaoChange(scaled_list)
		//fmt.Println(delta_tao)
		//For debugging

		if heuristic1 != 0 && heuristic1 != heuristic2 {
			scaled_list = rescale(scaled_list, alpha, beta, tao, gamma)
			heuristic1 = heuristic2
			heuristic2 = calculateHeuristic(delta_alpha, delta_beta, delta_gamma, delta_tao)
		} else {
			break
		}
		//fmt.Printf("%+v\n", scaled_list)
		//println(heuristic1)
	}
	var string_scaled_list [][]string
	for _, row := range scaled_list {
		var temp []string
		for _, item := range row {
			temp = append(temp, strconv.FormatFloat(item, 'g', 8, 64))
		}
		string_scaled_list = append(string_scaled_list, temp)
	}
	file, err := os.Create(*o)
	if err != nil {
		log.Println("Cannot create CSV file:", err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	err = writer.WriteAll(string_scaled_list)
	if err != nil {
		log.Println("Cannot write to CSV file:", err)
	}
}
