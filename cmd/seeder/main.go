package main

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ai-symptom-checker/config"
	"ai-symptom-checker/models"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	log.Println("Starting Database Seeder...")

	// Load env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or failed to load, proceeding with environment variables")
	}

	// Initialize Database connection
	config.Load()
	config.ConnectDB()
	db := config.DB

	// Migrate if not already migrated
	log.Println("Running AutoMigration for KnowledgeEntry...")
	err := db.AutoMigrate(&models.KnowledgeEntry{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	seedCoreDataset(db)
	seedHuggingFaceAPI(db)
	seedSymptom2DiseaseCSV(db)
	seedDerivedKnowledgeGraphCSV(db)

	log.Println("Seeding process finalized.")
	os.Exit(0)
}

func seedCoreDataset(db *gorm.DB) {
	log.Println("Seeding core clinical dataset...")
	entries := []models.KnowledgeEntry{
		{
			Title:          "Common Cold",
			Symptoms:       "Runny nose, Sore throat, Cough, Congestion, Slight body aches, Mild headache, Sneezing, Low-grade fever",
			Description:    "A viral infection of your nose and throat (upper respiratory tract). It's usually harmless.",
			Causes:         "Rhinoviruses are the most common cause, but many other viruses can cause a cold.",
			Advice:         "Get plenty of rest, drink plenty of fluids, and use over-the-counter cold medications to relieve symptoms. Seek medical attention if symptoms worsen or last longer than 10 days.",
			Source:         "Mayo Clinic",
			IsCoreDatasset: true,
			Status:         models.KnowledgeStatusActive,
		},
		{
			Title:          "Influenza (Flu)",
			Symptoms:       "Fever, Chills, Muscle aches, Cough, Congestion, Runny nose, Headaches, Fatigue",
			Description:    "A viral infection that attacks your respiratory system.",
			Causes:         "Influenza viruses traveling through the air in droplets.",
			Advice:         "Rest and hydration are key. Antiviral medications may be prescribed if caught early. Monitor fever closely.",
			Source:         "WHO",
			IsCoreDatasset: true,
			Status:         models.KnowledgeStatusActive,
		},
		{
			Title:          "Malaria",
			Symptoms:       "High fever, Shaking chills, Profuse sweating, Headache, Nausea, Vomiting, Abdominal pain, Diarrhea, Muscle or joint pain, Fatigue",
			Description:    "A disease caused by a parasite, transmitted by the bite of infected mosquitoes.",
			Causes:         "Plasmodium parasites transmitted by Anopheles mosquitoes.",
			Advice:         "**Urgent: Seek emergency medical care immediately.** Requires prescription antimalarial medication.",
			Source:         "WHO / NCDC",
			IsCoreDatasset: true,
			Status:         models.KnowledgeStatusActive,
		},
		{
			Title:          "Migraine",
			Symptoms:       "Throbbing headache, Sensitivity to light, Sensitivity to sound, Nausea, Vomiting, Visual aura",
			Description:    "A neurological condition that can cause intense, debilitating headaches.",
			Causes:         "Genetics, environmental factors, changes in the brainstem.",
			Advice:         "Rest in a quiet, dark room. Apply warm or cold compresses. Take over-the-counter pain relievers or prescribed medication. Stay hydrated.",
			Source:         "SymCat Guidelines",
			IsCoreDatasset: true,
			Status:         models.KnowledgeStatusActive,
		},
		{
			Title:          "COVID-19",
			Symptoms:       "Fever, Cough, Tiredness, Loss of taste or smell, Sore throat, Headache, Aches and pains, Diarrhea, Rash",
			Description:    "An infectious disease caused by the SARS-CoV-2 virus.",
			Causes:         "SARS-CoV-2 viral infection.",
			Advice:         "Isolate to prevent spreading. Rest and stay hydrated. Monitor oxygen levels. Seek immediate medical attention if experiencing difficulty breathing or chest pain.",
			Source:         "WHO",
			IsCoreDatasset: true,
			Status:         models.KnowledgeStatusActive,
		},
		{
			Title:          "Pregnancy-induced Headache",
			Symptoms:       "Dull aching pain, Throbbing pain, Sensitivity to light, Accompanied by nausea, Visual disturbances",
			Description:    "Headaches are common during pregnancy due to hormonal changes, increased blood volume, or posture.",
			Causes:         "Hormonal changes, increased blood volume, lack of sleep, low blood sugar, stress. Could indicate preeclampsia.",
			Advice:         "Rest, use a warm/cold compress, eat small regular meals, and drink plenty of water. Acetaminophen is generally safe, but ALWAYS consult your doctor. **Urgent**: If the headache is severe and accompanied by blurry vision, swelling, or right upper belly pain, go to the ER immediately (possible preeclampsia).",
			Source:         "Mayo Clinic",
			IsCoreDatasset: true,
			Status:         models.KnowledgeStatusActive,
		},
		{
			Title:          "Bacterial Meningitis",
			Symptoms:       "Sudden high fever, Stiff neck, Severe headache, Nausea or vomiting, Confusion, Seizures, Sleepiness, Sensitivity to light",
			Description:    "A serious infection of the meninges, the membranes that surround the brain and spinal cord.",
			Causes:         "Various bacteria including Streptococcus pneumoniae, Neisseria meningitidis.",
			Advice:         "**Urgent: SEEK IMMEDIATE EMERGENCY MEDICAL CARE.** Bacterial meningitis can be fatal in days without prompt antibiotic treatment.",
			Source:         "CDC Guidelines",
			IsCoreDatasset: true,
			Status:         models.KnowledgeStatusActive,
		},
		{
			Title:          "Asthma Exacerbation",
			Symptoms:       "Shortness of breath, Chest tightness or pain, Wheezing when exhaling, Trouble sleeping caused by shortness of breath, Coughing or wheezing attacks",
			Description:    "A condition in which your airways narrow and swell and may produce extra mucus.",
			Causes:         "Airborne allergens, respiratory infections, physical activity, cold air, air pollutants.",
			Advice:         "Use a quick-relief (rescue) inhaler. If symptoms do not improve rapidly or you have severe shortness of breath, **seek emergency medical care immediately.**",
			Source:         "NIH Clinical Tables",
			IsCoreDatasset: true,
			Status:         models.KnowledgeStatusActive,
		},
		{
			Title:          "Acute Appendicitis",
			Symptoms:       "Sudden pain on right side of lower abdomen, Pain that worsens if you cough or walk, Nausea and vomiting, Loss of appetite, Low-grade fever, Constipation or diarrhea, Abdominal bloating",
			Description:    "An inflammation of the appendix, a finger-shaped pouch that projects from your colon.",
			Causes:         "A blockage in the lining of the appendix that results in infection.",
			Advice:         "**Urgent: Go to the emergency room immediately.** Appendicitis requires prompt surgery to remove the appendix before it ruptures.",
			Source:         "Mayo Clinic",
			IsCoreDatasset: true,
			Status:         models.KnowledgeStatusActive,
		},
	}

	successCount := 0
	for _, entry := range entries {
		saveEntry(db, entry.Title, entry.Symptoms, entry.Description, entry.Causes, entry.Advice, entry.Source, true)
		successCount++
	}
	log.Printf("Core seeding checked %d entries.\n", successCount)
}

type HFResponse struct {
	Rows []HFRow `json:"rows"`
}

type HFRow struct {
	RowData HFRowData `json:"row"`
}

type HFRowData struct {
	Disease    string `json:"Disease"`
	Symptoms   string `json:"Symptoms"`
	Treatments string `json:"Treatments"`
}

func seedHuggingFaceAPI(db *gorm.DB) {
	url := "https://datasets-server.huggingface.co/rows?dataset=kamruzzaman-asif%2FDiseases_Dataset&config=default&split=QuyenAnh&offset=0&length=100"
	log.Println("Fetching HuggingFace dataset...")
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch HF dataset: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch HF dataset, status: %d\n", resp.StatusCode)
		return
	}

	var apiResp HFResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Printf("Failed to decode HF json: %v\n", err)
		return
	}

	count := 0
	for _, r := range apiResp.Rows {
		if r.RowData.Disease != "" {
			saveEntry(db, r.RowData.Disease, r.RowData.Symptoms, "Information loaded from ML disease dataset.", "", "Consult a healthcare professional for accurate treatment: "+r.RowData.Treatments, "HF: kamruzzaman-asif/Diseases_Dataset", false)
			count++
		}
	}
	log.Printf("Added/checked %d entries from HuggingFace dataset.\n", count)
}

func seedSymptom2DiseaseCSV(db *gorm.DB) {
	log.Println("Processing Symptom2Disease.csv...")
	// File paths options
	path := "C:/Users/Favour Opia/ReactDir/ai-symptoms-checker/Symptom2Disease.csv"

	f, err := os.Open(path)
	if err != nil {
		path = filepath.Join("..", "Symptom2Disease.csv")
		f, err = os.Open(path)
		if err != nil {
			log.Printf("Error opening Symptom2Disease.csv: %v\n", err)
			return
		}
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading Symptom2Disease CSV: %v\n", err)
		return
	}

	diseaseMap := make(map[string][]string)
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) >= 3 {
			label := strings.TrimSpace(row[1])
			text := strings.TrimSpace(row[2])
			if label != "" && text != "" {
				diseaseMap[label] = append(diseaseMap[label], text)
			}
		}
	}

	count := 0
	for label, texts := range diseaseMap {
		desc := strings.Join(texts, " | ")
		if len(desc) > 3000 {
			desc = desc[:2997] + "..."
		}
		saveEntry(db, label, desc, "Symptoms descriptions manually compiled from patient data.", "", "Consult a doctor for accurate diagnosis.", "Symptom2Disease.csv", false)
		count++
	}
	log.Printf("Added/checked %d unique diseases from Symptom2Disease.csv.\n", count)
}

func seedDerivedKnowledgeGraphCSV(db *gorm.DB) {
	log.Println("Processing DerivedKnowledgeGraph_final.csv...")
	path := "C:/Users/Favour Opia/ReactDir/ai-symptoms-checker/DerivedKnowledgeGraph_final.csv"

	f, err := os.Open(path)
	if err != nil {
		path = filepath.Join("..", "DerivedKnowledgeGraph_final.csv")
		f, err = os.Open(path)
		if err != nil {
			log.Printf("Error opening KnowledgeGraph CSV: %v\n", err)
			return
		}
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading KnowledgeGraph CSV: %v\n", err)
		return
	}

	re := regexp.MustCompile(`\s*\([0-9.]+\)`)
	count := 0
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) >= 2 {
			disease := strings.TrimSpace(row[0])
			symptomsRaw := strings.TrimSpace(row[1])
			if disease != "" && symptomsRaw != "" {
				symptomsClean := re.ReplaceAllString(symptomsRaw, "")
				saveEntry(db, disease, symptomsClean, "Derived from symptom knowledge graph associations.", "", "Consult a healthcare professional for accurate treatment.", "DerivedKnowledgeGraph_final.csv", false)
				count++
			}
		}
	}
	log.Printf("Added/checked %d diseases from Knowledge Graph.\n", count)
}

func saveEntry(db *gorm.DB, title, symptoms, desc, causes, advice, source string, isCore bool) {
	if len(title) > 0 {
		title = strings.ToUpper(title[:1]) + title[1:]
	}

	var existing models.KnowledgeEntry
	res := db.Where("disease_name = ?", title).First(&existing)
	if res.RowsAffected == 0 {
		entry := models.KnowledgeEntry{
			Title:          title,
			Symptoms:       symptoms,
			Description:    desc,
			Causes:         causes,
			Advice:         advice,
			Source:         source,
			IsCoreDatasset: isCore,
			Status:         models.KnowledgeStatusActive,
		}
		if err := db.Create(&entry).Error; err != nil {
			log.Printf("Error inserting %s: %v\n", title, err)
		}
	} else {
		// optionally update existing
	}
}
