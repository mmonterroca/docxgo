using System;
using System.IO;
using System.Linq;
using DocumentFormat.OpenXml.Packaging;
using DocumentFormat.OpenXml.Validation;

var defaultPath = Path.Combine("examples", "07_advanced", "07_advanced_demo.docx");
var targetPath = args.Length > 0 ? args[0] : defaultPath;

Console.WriteLine($"Validating: {targetPath}");

if (!File.Exists(targetPath))
{
	Console.Error.WriteLine($"File not found: {Path.GetFullPath(targetPath)}");
	return 1;
}

try
{
	using var document = WordprocessingDocument.Open(targetPath, false);
	var validator = new OpenXmlValidator();
	var errors = validator.Validate(document).ToList();

	if (errors.Count == 0)
	{
		Console.WriteLine("Open XML validation passed with no errors.");
		return 0;
	}

	Console.WriteLine($"Open XML validation found {errors.Count} error(s):");

	for (var i = 0; i < errors.Count; i++)
	{
		var error = errors[i];
		var part = error.Part?.Uri?.ToString() ?? "(unknown part)";
		var path = error.Path?.XPath ?? "(no XPath provided)";

		Console.WriteLine($"{i + 1}. {error.Description}");
		Console.WriteLine($"   Part: {part}");
		Console.WriteLine($"   Path: {path}");
	Console.WriteLine($"   Error Type: {error.ErrorType}");
	}

	return 1;
}
catch (Exception ex)
{
	Console.Error.WriteLine($"Validation failed: {ex.Message}");
	return 2;
}
