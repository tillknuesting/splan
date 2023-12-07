package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Teacher struct {
	ID        string
	Name      string
	Subjects  []string       // Subjects that the teacher can teach
	Available []time.Weekday // Days available to teach
}

type Room struct {
	ID       string
	Capacity int
}

type TimeSlot struct {
	Day   time.Weekday
	Start time.Time
	End   time.Time
}

type Class struct {
	Subject  string
	Teacher  *Teacher
	Room     *Room
	TimeSlot *TimeSlot
	Capacity int
}

// Check if the time slots overlap
func timeSlotsOverlap(slot1, slot2 *TimeSlot) bool {
	return slot1.Day == slot2.Day && slot1.Start.Before(slot2.End) && slot2.Start.Before(slot1.End)
}

// Check if the teacher is available to teach at the given time slot
func checkTeacherAvailability(teacher *Teacher, timeSlot *TimeSlot) bool {
	// Check if the teacher is available on the day of the time slot
	for _, availableDay := range teacher.Available {
		if availableDay == timeSlot.Day {
			// Further refinement for specific hours can be added here if needed
			return true
		}
	}
	return false
}

// Check if the room is available at the given time slot
// This function requires access to all classes to check room allocation
func checkRoomAvailability(classes []*Class, room *Room, timeSlot *TimeSlot) bool {
	for _, class := range classes {
		if class.Room.ID == room.ID && class.TimeSlot == timeSlot {
			return false
		}
	}
	return true
}

// Check if the class capacity fits the room capacity
func checkRoomCapacity(class *Class, room *Room) bool {
	return class.Capacity <= room.Capacity
}

// Check if the teacher is qualified to teach the subject
func checkTeacherQualification(teacher *Teacher, subject string) bool {
	for _, subj := range teacher.Subjects {
		if subj == subject {
			return true
		}
	}
	return false
}

type Gene struct {
	ClassAssignment *Class
}

type Chromosome struct {
	Genes []Gene
}

type Population struct {
	Timetables []Chromosome
}

// Initialize a random timetable (for the initial population)
func initializeRandomTimetable(classes []*Class, teachers []*Teacher, rooms []*Room, timeSlots []*TimeSlot) Chromosome {
	var timetable Chromosome
	for _, class := range classes {
		// Randomly assign a teacher, room, and time slot
		assignedTeacher := teachers[rand.Intn(len(teachers))]
		assignedRoom := rooms[rand.Intn(len(rooms))]
		assignedTimeSlot := timeSlots[rand.Intn(len(timeSlots))]

		// Create a gene with the random assignment
		gene := Gene{
			ClassAssignment: &Class{
				Subject:  class.Subject,
				Teacher:  assignedTeacher,
				Room:     assignedRoom,
				TimeSlot: assignedTimeSlot,
				Capacity: class.Capacity,
			},
		}
		timetable.Genes = append(timetable.Genes, gene)
	}
	return timetable
}

// Calculate the fitness score of a timetable
func calculateFitness(chromosome Chromosome, classes []*Class) int {
	fitness := 0

	// Check for teacher and room conflicts, teacher qualifications, and teacher availability
	for i, gene1 := range chromosome.Genes {
		for j, gene2 := range chromosome.Genes {
			if i != j {
				if timeSlotsOverlap(gene1.ClassAssignment.TimeSlot, gene2.ClassAssignment.TimeSlot) {
					if gene1.ClassAssignment.Teacher.ID == gene2.ClassAssignment.Teacher.ID {
						fitness -= 10 // Significantly penalize teacher conflict
					}
					if gene1.ClassAssignment.Room.ID == gene2.ClassAssignment.Room.ID {
						fitness-- // Room conflict
					}
				}
			}
		}

		if !checkTeacherQualification(gene1.ClassAssignment.Teacher, gene1.ClassAssignment.Subject) {
			fitness-- // Teacher not qualified
		}

		if !checkRoomCapacity(gene1.ClassAssignment, gene1.ClassAssignment.Room) {
			fitness-- // Room capacity exceeded
		}

		if !checkTeacherAvailability(gene1.ClassAssignment.Teacher, gene1.ClassAssignment.TimeSlot) {
			fitness-- // Teacher not available
		}
	}

	return fitness
}

// TournamentSelection selects the best individual from a randomly chosen subset
func TournamentSelection(population Population, tournamentSize int, classes []*Class) Chromosome {
	best := -1
	bestFitness := -1000 // Start with a very low fitness

	for i := 0; i < tournamentSize; i++ {
		individualIndex := rand.Intn(len(population.Timetables))
		currentFitness := calculateFitness(population.Timetables[individualIndex], classes)
		if best == -1 || currentFitness > bestFitness {
			best = individualIndex
			bestFitness = currentFitness
		}
	}
	return population.Timetables[best]
}

// CreateNewGeneration creates a new generation using tournament selection, crossover, and mutation
func CreateNewGeneration(population Population, tournamentSize int, populationSize int, classes []*Class, teachers []*Teacher, rooms []*Room, timeSlots []*TimeSlot, mutationRate float64) Population {
	var newGeneration Population

	for i := 0; i < populationSize; i += 2 {
		parent1 := TournamentSelection(population, tournamentSize, classes)
		parent2 := TournamentSelection(population, tournamentSize, classes)

		child1 := crossover(parent1, parent2)
		child2 := crossover(parent2, parent1)

		// Apply mutation
		mutatedChild1 := mutate(child1, teachers, rooms, timeSlots, mutationRate)
		mutatedChild2 := mutate(child2, teachers, rooms, timeSlots, mutationRate)

		newGeneration.Timetables = append(newGeneration.Timetables, mutatedChild1)
		if len(newGeneration.Timetables) < populationSize {
			newGeneration.Timetables = append(newGeneration.Timetables, mutatedChild2)
		}
	}

	return newGeneration
}

// Perform one-point crossover between two timetables
func crossover(parent1, parent2 Chromosome) Chromosome {
	crossoverPoint := rand.Intn(len(parent1.Genes))
	var childGenes []Gene

	for i := 0; i < len(parent1.Genes); i++ {
		if i < crossoverPoint {
			childGenes = append(childGenes, parent1.Genes[i])
		} else {
			childGenes = append(childGenes, parent2.Genes[i])
		}
	}

	return Chromosome{Genes: childGenes}
}

// Mutate a single timetable (chromosome) by randomly altering its genes
func mutate(chromosome Chromosome, teachers []*Teacher, rooms []*Room, timeSlots []*TimeSlot, mutationRate float64) Chromosome {
	for i := 0; i < len(chromosome.Genes); i++ {
		if rand.Float64() < mutationRate {
			// Randomly mutate teacher, room, or time slot
			mutationChoice := rand.Intn(3)
			switch mutationChoice {
			case 0: // Mutate teacher
				chromosome.Genes[i].ClassAssignment.Teacher = teachers[rand.Intn(len(teachers))]
			case 1: // Mutate room
				chromosome.Genes[i].ClassAssignment.Room = rooms[rand.Intn(len(rooms))]
			case 2: // Mutate time slot
				chromosome.Genes[i].ClassAssignment.TimeSlot = timeSlots[rand.Intn(len(timeSlots))]
			}
		}
	}
	return chromosome
}

func PrintTimetable(timetable Chromosome) {
	// Define the structure for the timetable
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	times := []string{"08:00-10:00", "10:30-11:30", "11:30-12:30", "13:30-14:30", "14:30-15:30"}
	grid := make(map[string]map[string][]string) // Use slices of strings for multiline cells

	// Initialize grid with empty values
	for _, day := range days {
		grid[day] = make(map[string][]string)
		for _, time := range times {
			grid[day][time] = []string{"Free"}
		}
	}

	// Populate the grid with class details
	for _, gene := range timetable.Genes {
		class := gene.ClassAssignment
		day := class.TimeSlot.Day.String()
		timeRange := class.TimeSlot.Start.Format("15:04") + "-" + class.TimeSlot.End.Format("15:04")
		details := class.Subject + " (T: " + class.Teacher.Name + ", R: " + class.Room.ID + ")"
		grid[day][timeRange] = splitIntoLines(details, 19) // Split details into lines of up to 19 characters
	}

	// Print the timetable in a grid format
	fmt.Println("\nTimetable:")
	printLine := func() { fmt.Println(strings.Repeat("-", 107)) }
	printLine()

	// Print the header row
	fmt.Printf("| %-10s |", "Time/Day")
	for _, day := range days {
		fmt.Printf(" %-19s |", day)
	}
	fmt.Println()
	printLine()

	// Print the timetable rows
	for _, time := range times {
		maxLines := getMaxLines(grid, time, days)
		for i := 0; i < maxLines; i++ {
			if i == 0 {
				fmt.Printf("| %-10s |", time) // Print the time slot only on the first line
			} else {
				fmt.Printf("| %-10s |", "") // Empty space for subsequent lines
			}
			for _, day := range days {
				if i < len(grid[day][time]) {
					fmt.Printf(" %-19s |", grid[day][time][i])
				} else {
					fmt.Printf(" %-19s |", "") // Fill empty space if no more lines
				}
			}
			fmt.Println()
		}
		printLine()
	}
}

// splitIntoLines splits a string into lines of maximum specified length
func splitIntoLines(s string, maxLen int) []string {
	var lines []string
	for len(s) > 0 {
		if len(s) > maxLen {
			lines = append(lines, s[:maxLen])
			s = s[maxLen:]
		} else {
			lines = append(lines, s)
			break
		}
	}
	return lines
}

// getMaxLines returns the maximum number of lines needed for a given time slot across all days
func getMaxLines(grid map[string]map[string][]string, time string, days []string) int {
	maxLines := 0
	for _, day := range days {
		if len(grid[day][time]) > maxLines {
			maxLines = len(grid[day][time])
		}
	}
	return maxLines
}

func main() {
	// Sample Teachers
	teachers := []*Teacher{
		{
			ID:        "T1",
			Name:      "Mr. Smith",
			Subjects:  []string{"Mathematics", "Physics"},
			Available: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
		},
		{
			ID:        "T2",
			Name:      "Ms. Johnson",
			Subjects:  []string{"History", "English"},
			Available: []time.Weekday{time.Tuesday, time.Thursday},
		},
		{
			ID:        "T3",
			Name:      "Mr. Williams",
			Subjects:  []string{"English", "Literature"},
			Available: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		},
		{
			ID:        "T4",
			Name:      "Ms. Brown",
			Subjects:  []string{"Chemistry", "Biology"},
			Available: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		},
		{
			ID:        "T5",
			Name:      "Mr. Green",
			Subjects:  []string{"Physical Education", "Health"},
			Available: []time.Weekday{time.Tuesday, time.Thursday, time.Friday},
		},
		{
			ID:        "T6",
			Name:      "Ms. Davis",
			Subjects:  []string{"Art", "Music"},
			Available: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
		},
		{
			ID:        "T7",
			Name:      "Mr. Wilson",
			Subjects:  []string{"Computer Science", "Mathematics"},
			Available: []time.Weekday{time.Tuesday, time.Thursday},
		},
		{
			ID:        "T8",
			Name:      "Ms. Taylor",
			Subjects:  []string{"Foreign Language", "Geography"},
			Available: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
		},
		// Additional teachers can be added for more subject variety or to cover any gaps
	}

	// Sample Rooms
	rooms := []*Room{
		{ID: "R101", Capacity: 30},
		{ID: "R102", Capacity: 30},
		{ID: "R103", Capacity: 30},
		{ID: "R104", Capacity: 30},
		{ID: "R105", Capacity: 30},
		{ID: "R106", Capacity: 30},
		{ID: "R107", Capacity: 30},
		{ID: "R108", Capacity: 30},
		{ID: "R109", Capacity: 30},
	}

	// Adjusted Time Slots
	timeSlots := []*TimeSlot{
		// Monday
		{Day: time.Monday, Start: time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC), End: time.Date(0, 0, 0, 10, 0, 0, 0, time.UTC)},
		{Day: time.Monday, Start: time.Date(0, 0, 0, 10, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 11, 30, 0, 0, time.UTC)},
		{Day: time.Monday, Start: time.Date(0, 0, 0, 11, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 12, 30, 0, 0, time.UTC)},
		{Day: time.Monday, Start: time.Date(0, 0, 0, 13, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC)},
		{Day: time.Monday, Start: time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 15, 30, 0, 0, time.UTC)},

		// Tuesday
		{Day: time.Tuesday, Start: time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC), End: time.Date(0, 0, 0, 10, 0, 0, 0, time.UTC)},
		{Day: time.Tuesday, Start: time.Date(0, 0, 0, 10, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 11, 30, 0, 0, time.UTC)},
		{Day: time.Tuesday, Start: time.Date(0, 0, 0, 11, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 12, 30, 0, 0, time.UTC)},
		{Day: time.Tuesday, Start: time.Date(0, 0, 0, 13, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC)},
		{Day: time.Tuesday, Start: time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 15, 30, 0, 0, time.UTC)},

		// Wednesday
		{Day: time.Wednesday, Start: time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC), End: time.Date(0, 0, 0, 10, 0, 0, 0, time.UTC)},
		{Day: time.Wednesday, Start: time.Date(0, 0, 0, 10, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 11, 30, 0, 0, time.UTC)},
		{Day: time.Wednesday, Start: time.Date(0, 0, 0, 11, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 12, 30, 0, 0, time.UTC)},
		{Day: time.Wednesday, Start: time.Date(0, 0, 0, 13, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC)},
		{Day: time.Wednesday, Start: time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 15, 30, 0, 0, time.UTC)},

		// Thursday
		{Day: time.Thursday, Start: time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC), End: time.Date(0, 0, 0, 10, 0, 0, 0, time.UTC)},
		{Day: time.Thursday, Start: time.Date(0, 0, 0, 10, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 11, 30, 0, 0, time.UTC)},
		{Day: time.Thursday, Start: time.Date(0, 0, 0, 11, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 12, 30, 0, 0, time.UTC)},
		{Day: time.Thursday, Start: time.Date(0, 0, 0, 13, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC)},
		{Day: time.Thursday, Start: time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 15, 30, 0, 0, time.UTC)},

		// Friday
		{Day: time.Friday, Start: time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC), End: time.Date(0, 0, 0, 10, 0, 0, 0, time.UTC)},
		{Day: time.Friday, Start: time.Date(0, 0, 0, 10, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 11, 30, 0, 0, time.UTC)},
		{Day: time.Friday, Start: time.Date(0, 0, 0, 11, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 12, 30, 0, 0, time.UTC)},
		{Day: time.Friday, Start: time.Date(0, 0, 0, 13, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC)},
		{Day: time.Friday, Start: time.Date(0, 0, 0, 14, 30, 0, 0, time.UTC), End: time.Date(0, 0, 0, 15, 30, 0, 0, time.UTC)},
	}

	// Sample Classes (initially without assignments)
	classes := []*Class{
		{Subject: "Mathematics"},
		{Subject: "History"},
		{Subject: "Physics"},
		{Subject: "English"},
		{Subject: "Biology"},
		{Subject: "Chemistry"},
		{Subject: "Computer Science"},
		{Subject: "Physical Education"},
		{Subject: "Art"},
		{Subject: "Music"},
		{Subject: "Foreign Language"}, // You can specify particular languages like Spanish, French, etc.
		{Subject: "Geography"},
		{Subject: "Literature"},
		// Add more classes as needed
	}

	// Random seed for random number generation
	rand.Seed(time.Now().UnixNano())

	populationSize := 10 // Maintaining a constant population size

	// Initialize a population of timetables
	var population Population
	for i := 0; i < populationSize; i++ { // Example: population size of 10
		population.Timetables = append(population.Timetables, initializeRandomTimetable(classes, teachers, rooms, timeSlots))
	}

	// Calculate and display the fitness of each timetable in the population
	for i, timetable := range population.Timetables {
		fitness := calculateFitness(timetable, classes)
		fmt.Printf("Timetable %d: Fitness = %d\n", i+1, fitness)
		if fitness == 0 {
			fmt.Println("Timetable is valid!")
			return
		}
	}

	// Define the number of generations for the GA to run
	numGenerations := 25

	// Parameters for the genetic algorithm
	tournamentSize := 5 // Example: size of tournament for selection

	mutationRate := 0.05 // For example, 5% mutation rate

	// Genetic Algorithm Loop
	for generation := 0; generation < numGenerations; generation++ {
		// Selection, Crossover, and Mutation
		population = CreateNewGeneration(population, tournamentSize, populationSize, classes, teachers, rooms, timeSlots, mutationRate)

		// Initialize variable to track the best fitness in this generation
		bestFitnessInGeneration := -100000000

		// Evaluate the new generation
		for _, timetable := range population.Timetables {
			currentFitness := calculateFitness(timetable, classes)

			// Check if current timetable has the best fitness so far in this generation
			if currentFitness > bestFitnessInGeneration {
				bestFitnessInGeneration = currentFitness
			}
		}

		// Print the best fitness of this generation
		fmt.Printf("Generation %d: Best Fitness = %d\n", generation+1, bestFitnessInGeneration)
	}

	bestTimetableIndex := 0
	bestFitness := -100000000
	for i, timetable := range population.Timetables {
		currentFitness := calculateFitness(timetable, classes)
		fmt.Println("Fitness of Timetable", i+1, ":", currentFitness)
		if currentFitness > bestFitness {
			fmt.Println("New best fitness found:", currentFitness)
			bestFitness = currentFitness
			bestTimetableIndex = i
		}
	}

	fmt.Println("Best Timetable:")
	PrintTimetable(population.Timetables[bestTimetableIndex])

}
