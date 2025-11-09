package usernamegenerator

import (
	"fmt"
	"math/rand"
)

var Adjectives = []string{
	"Happy", "Clever", "Swift", "Brave", "Calm", "Proud", "Witty", "Gentle",
	"Bright", "Fierce", "Lucky", "Magic", "Noble", "Quick", "Royal", "Sharp",
	"Super", "Tough", "Vivid", "Wild", "Zesty", "Alpha", "Beta", "Cosmic",
	"Digital", "Epic", "Fusion", "Hyper", "Quantum", "Stealth", "Agile",
	"Bold", "Crisp", "Daring", "Eager", "Fancy", "Golden", "Heroic", "Icy",
	"Jolly", "Keen", "Lively", "Mighty", "Neon", "Oceanic", "Peaceful", "Quiet",
	"Radiant", "Silent", "Titanic", "Unique", "Vibrant", "Wonder", "Young",
	"Zealous", "Atomic", "Blazing", "Crystal", "Dynamic", "Electric", "Fiery",
	"Glowing", "Harmonic", "Infinite", "Jade", "Lunar", "Mystic", "Nova",
	"Orbital", "Primal", "Quantum", "Rapid", "Solar", "Thunder", "Ultra",
	"Vortex", "Wind", "Xenon", "Yellow", "Zen", "Amber", "Bronze", "Copper",
	"Diamond", "Emerald", "Forest", "Graphite", "Honey", "Iron", "Jasmine",
	"Kiwi", "Lavender", "Marble", "Nickel", "Opal", "Pearl", "Quartz", "Ruby",
	"Sapphire", "Topaz", "Umber", "Velvet", "Willow", "Xylon", "Yarrow", "Zircon",
}

var Animals = []string{
	"Panda", "Tiger", "Eagle", "Shark", "Wolf", "Fox", "Bear", "Owl",
	"Lion", "Falcon", "Hawk", "Bison", "Cobra", "Dragon", "Fox", "Gecko",
	"Hound", "Jaguar", "Koala", "Lynx", "Mouse", "Otter", "Panther", "Quail",
	"Raven", "Sloth", "Turtle", "Viper", "Whale", "Zebra", "Alligator",
	"Badger", "Cheetah", "Dolphin", "Elephant", "Flamingo", "Gorilla", "Hamster",
	"Iguana", "Jackal", "Kangaroo", "Leopard", "Mongoose", "Narwhal", "Ocelot",
	"Penguin", "Quokka", "Raccoon", "Seal", "Toucan", "Urchin", "Vulture",
	"Wombat", "Yak", "Alpaca", "Beaver", "Chameleon", "Donkey", "Ferret",
	"Gazelle", "Heron", "Impala", "Jellyfish", "Kudu", "Lemur", "Manatee",
	"Newt", "Octopus", "Parrot", "Quail", "Robin", "Squirrel", "Tapir",
	"Unicorn", "Vole", "Walrus", "Xerus", "Yeti", "Antelope", "Bobcat",
	"Caribou", "Dingo", "Elk", "Frog", "Goat", "Hyena", "Ibis", "Jackrabbit",
	"Kingfisher", "Lobster", "Macaw", "Nightingale", "Orca", "Peacock",
	"Rabbit", "Salmon", "Trout", "Urchin", "Weasel", "Xrayfish", "Yellowjacket",
	"Zebu", "Armadillo", "Butterfly", "Crab", "Duck", "Emu", "Finch", "Gull",
}

func GenerateRandomUsername() string {
	adj := Adjectives[rand.Intn(len(Adjectives))]
	animal := Animals[rand.Intn(len(Animals))]
	digits := rand.Intn(100000)
	return fmt.Sprintf("%s_%s_%05d", adj, animal, digits)
}
