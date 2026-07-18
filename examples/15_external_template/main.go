// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

// Example 15: External Template Mail Merge
//
// This example demonstrates how to use an existing Word document as a mail merge
// template. It reads a real .docx file created in Microsoft Word that contains
// MERGEFIELD fields with «guillemet» delimiters, extracts the placeholders,
// validates the data, and produces merged output documents.
//
// This showcases the library's ability to:
// - Read existing .docx templates created in Word, Google Docs, etc.
// - Use custom delimiters (« ») instead of the default {{ }}
// - Detect and list all placeholders in the template
// - Validate data completeness before merging
// - Batch-merge a single template into multiple personalized documents
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/pkg/template"
)

func main() {
	// The template file was created in Microsoft Word with MERGEFIELD fields.
	// It uses «placeholder» syntax (guillemet delimiters).
	templatePath := "reminder_letter.docx"

	// -------------------------------------------------------------------------
	// Step 1: Open the external Word template
	// -------------------------------------------------------------------------
	fmt.Println("Step 1: Opening external Word template...")
	doc, err := docx.OpenDocument(templatePath)
	if err != nil {
		log.Fatalf("Failed to open template: %v", err)
	}
	fmt.Printf("  Opened: %s\n\n", templatePath)

	// -------------------------------------------------------------------------
	// Step 2: Configure custom delimiters for Word-style merge fields
	// -------------------------------------------------------------------------
	fmt.Println("Step 2: Configuring custom delimiters for Word-style merge fields...")
	// Microsoft Word uses « » (guillemet) delimiters for mail merge fields.
	// Our template engine supports custom delimiters through MergeOptions.
	opts := template.MergeOptions{
		OpenDelimiter:  "«",
		CloseDelimiter: "»",
		StrictMode:     true, // require all placeholders to have data
	}
	fmt.Println("  Delimiters: « »")
	fmt.Println("  Strict mode: enabled")
	fmt.Println()

	// -------------------------------------------------------------------------
	// Step 3: Inspect the template — discover all placeholders
	// -------------------------------------------------------------------------
	fmt.Println("Step 3: Inspecting template placeholders...")

	// Use FindPlaceholdersCustom with a pattern matching « » delimiters
	pattern := buildGuillemetsPattern()
	placeholders := template.FindPlaceholdersCustom(doc, pattern)

	// Deduplicate names
	seen := make(map[string]bool)
	var uniqueNames []string
	for _, p := range placeholders {
		if !seen[p.Name] {
			seen[p.Name] = true
			uniqueNames = append(uniqueNames, p.Name)
		}
	}

	fmt.Printf("  Found %d unique placeholders (%d total occurrences):\n", len(uniqueNames), len(placeholders))
	for _, name := range uniqueNames {
		count := 0
		for _, p := range placeholders {
			if p.Name == name {
				count++
			}
		}
		if count > 1 {
			fmt.Printf("    - «%s» (appears %d times)\n", name, count)
		} else {
			fmt.Printf("    - «%s»\n", name)
		}
	}
	fmt.Println()

	// -------------------------------------------------------------------------
	// Step 4: Prepare merge data and validate
	// -------------------------------------------------------------------------
	fmt.Println("Step 4: Validating template against data...")

	data := template.MergeData{
		// Organization info
		"Org_Name":       "TechForward Solutions",
		"Org_Address":    "456 Innovation Drive, Suite 200",
		"Org_City":       "Austin",
		"Org_State":      "TX",
		"Org_PostalCode": "78701",

		// Date
		"Today": time.Now().Format("January 2, 2006"),

		// Contact info
		"Contact_FullName":          "Sarah Mitchell",
		"Contact_Title":            "VP of Sales",
		"Account_Name":             "Global Enterprises Inc.",
		"Contact_MailingAddress":   "789 Commerce Blvd",
		"Contact_MailingCity":      "Chicago",
		"Contact_MailingState":     "IL",
		"Contact_MailingPostalCode": "60601",
		"Contact_FirstName":        "Sarah",

		// User (sender) info
		"User_FullName": "James Rodriguez",
		"User_Title":    "Director of Training",
		"User_Company":  "TechForward Solutions",
		"User_Phone":    "(512) 555-0147",
		"User_Fax":      "(512) 555-0148",
		"User_Email":    "j.rodriguez@techforward.com",
	}

	validationErrors := template.ValidateTemplate(doc, data, opts)
	hasErrors := false
	for _, ve := range validationErrors {
		if ve.Severity == template.SeverityError {
			fmt.Printf("  ERROR: %s\n", ve)
			hasErrors = true
		} else {
			fmt.Printf("  WARNING: %s\n", ve)
		}
	}
	if hasErrors {
		log.Fatal("Template validation failed — aborting merge")
	}
	if len(validationErrors) == 0 {
		fmt.Println("  All placeholders have matching data keys!")
	}
	fmt.Println()

	// -------------------------------------------------------------------------
	// Step 5: Merge the template to produce a personalized letter
	// -------------------------------------------------------------------------
	fmt.Println("Step 5: Merging template with data...")

	if err := template.MergeTemplate(doc, data, opts); err != nil {
		log.Fatalf("Merge failed: %v", err)
	}

	outputPath := "reminder_sarah_mitchell.docx"
	if err := doc.SaveAs(outputPath); err != nil {
		log.Fatalf("Failed to save merged document: %v", err)
	}
	fmt.Printf("  Saved: %s\n\n", outputPath)

	// -------------------------------------------------------------------------
	// Step 6: Verify — reopen the merged document and print its content
	// -------------------------------------------------------------------------
	fmt.Println("Step 6: Verifying merged document content...")

	merged, err := docx.OpenDocument(outputPath)
	if err != nil {
		log.Fatalf("Failed to reopen merged document: %v", err)
	}

	for _, p := range merged.Paragraphs() {
		text := p.Text()
		if text != "" {
			fmt.Printf("  %s\n", text)
		}
	}
	fmt.Println()

	// -------------------------------------------------------------------------
	// Step 7: Batch merge — send reminders to multiple contacts
	// -------------------------------------------------------------------------
	fmt.Println("Step 7: Batch merge for multiple contacts...")

	contacts := []template.MergeData{
		{
			"Org_Name": "TechForward Solutions", "Org_Address": "456 Innovation Drive, Suite 200",
			"Org_City": "Austin", "Org_State": "TX", "Org_PostalCode": "78701",
			"Today":                    time.Now().Format("January 2, 2006"),
			"Contact_FullName":         "Michael Chen",
			"Contact_Title":            "Sales Director",
			"Account_Name":             "Pacific Rim Trading Co.",
			"Contact_MailingAddress":    "100 Harbor View",
			"Contact_MailingCity":       "San Francisco",
			"Contact_MailingState":      "CA",
			"Contact_MailingPostalCode": "94105",
			"Contact_FirstName":         "Michael",
			"User_FullName": "James Rodriguez", "User_Title": "Director of Training",
			"User_Company": "TechForward Solutions", "User_Phone": "(512) 555-0147",
			"User_Fax": "(512) 555-0148", "User_Email": "j.rodriguez@techforward.com",
		},
		{
			"Org_Name": "TechForward Solutions", "Org_Address": "456 Innovation Drive, Suite 200",
			"Org_City": "Austin", "Org_State": "TX", "Org_PostalCode": "78701",
			"Today":                    time.Now().Format("January 2, 2006"),
			"Contact_FullName":         "Emily Watson",
			"Contact_Title":            "Regional Sales Manager",
			"Account_Name":             "Midwest Solutions Group",
			"Contact_MailingAddress":    "2500 Lake Shore Drive",
			"Contact_MailingCity":       "Minneapolis",
			"Contact_MailingState":      "MN",
			"Contact_MailingPostalCode": "55401",
			"Contact_FirstName":         "Emily",
			"User_FullName": "James Rodriguez", "User_Title": "Director of Training",
			"User_Company": "TechForward Solutions", "User_Phone": "(512) 555-0147",
			"User_Fax": "(512) 555-0148", "User_Email": "j.rodriguez@techforward.com",
		},
	}

	for _, contactData := range contacts {
		// Reopen the template for each recipient (merge modifies in-place)
		contactDoc, err := docx.OpenDocument(templatePath)
		if err != nil {
			log.Fatalf("Failed to open template: %v", err)
		}

		if err := template.MergeTemplate(contactDoc, contactData, opts); err != nil {
			log.Fatalf("Merge failed for %s: %v", contactData["Contact_FullName"], err)
		}

		// Clean filename from contact name
		name := contactData["Contact_FullName"]
		outFile := fmt.Sprintf("reminder_%s.docx", sanitizeFilename(name))
		if err := contactDoc.SaveAs(outFile); err != nil {
			log.Fatalf("Failed to save: %v", err)
		}
		fmt.Printf("  Saved: %s\n", outFile)
	}

	fmt.Println("\nDone! All reminder letters generated successfully.")

	// Cleanup generated files (optional — comment out to keep them)
	// cleanup(outputPath, contacts)
}

// buildGuillemetsPattern creates a regex pattern for « » delimited placeholders.
func buildGuillemetsPattern() *regexp.Regexp {
	return regexp.MustCompile(`«(\w+(?:\.\w+)*)»`)
}

// sanitizeFilename replaces spaces with underscores for safe filenames.
func sanitizeFilename(name string) string {
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		if name[i] == ' ' {
			result = append(result, '_')
		} else {
			result = append(result, name[i])
		}
	}
	return string(result)
}

// cleanup removes generated .docx files to keep the example directory clean.
// It preserves the original template file (reminder_letter.docx).
func cleanup(singleOutput string, contacts []template.MergeData) {
	os.Remove(singleOutput)
	for _, c := range contacts {
		name := c["Contact_FullName"]
		os.Remove(fmt.Sprintf("reminder_%s.docx", sanitizeFilename(name)))
	}

	// Also remove any other generated files, but preserve the template
	matches, _ := filepath.Glob("reminder_*.docx")
	for _, m := range matches {
		base := filepath.Base(m)
		if base == "reminder_letter.docx" {
			continue // preserve the original template
		}
		os.Remove(m)
	}
}
