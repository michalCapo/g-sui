package pages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/michalCapo/g-sui/ui"
)

type imageUploadData struct {
	Title       string `validate:"required"`
	Description string
	ImageFile   string // Not used for file upload, just for display
}

func ImageUploadSubmit(ctx *ui.Context) string {
	form := imageUploadData{}
	err := ctx.Body(&form)
	if err != nil {
		return renderImageUploadForm(ctx, &form, &err)
	}

	// Validate form fields
	v := validator.New()
	if err := v.Struct(form); err != nil {
		return renderImageUploadForm(ctx, &form, &err)
	}

	// Handle file upload using new ctx.File() method
	file, err := ctx.File("image")
	if err != nil {
		ctx.Error("Failed to process file: " + err.Error())
		return renderImageUploadForm(ctx, &form, nil)
	}
	if file == nil {
		ctx.Error("No file uploaded")
		return renderImageUploadForm(ctx, &form, nil)
	}

	// Validate file type
	if !strings.HasPrefix(file.ContentType, "image/") {
		ctx.Error("File must be an image")
		return renderImageUploadForm(ctx, &form, nil)
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		ctx.Error("Image size must be less than 5MB")
		return renderImageUploadForm(ctx, &form, nil)
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "/tmp/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		ctx.Error("Failed to create upload directory: " + err.Error())
		return renderImageUploadForm(ctx, &form, nil)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Name)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%d%s", os.Getpid(), ext)
	filepath := filepath.Join(uploadDir, filename)

	// Save file
	if err := os.WriteFile(filepath, file.Data, 0644); err != nil {
		ctx.Error("Failed to save file: " + err.Error())
		return renderImageUploadForm(ctx, &form, nil)
	}

	ctx.Success(fmt.Sprintf("Image uploaded successfully! File: %s (Size: %.2f KB)", file.Name, float64(file.Size)/1024))

	// Reset form
	form = imageUploadData{}
	return renderImageUploadForm(ctx, &form, nil)
}

func renderImageUploadForm(ctx *ui.Context, data *imageUploadData, err *error) string {
	target := ui.Target()
	form := ui.FormNew(ctx.Submit(ImageUploadSubmit).Replace(target))

	result := "Upload an image to see the result here."

	if err == nil && data.Title != "" {
		result = fmt.Sprintf("Form submitted with title: %s", data.Title)
	}

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4", target)(
		ui.Div("text-2xl font-bold")("Image Upload Form"),
		ui.Div("text-gray-600")("Demonstrates file upload with image preview and validation."),

		ui.Card().Header("<h3 class='font-bold text-lg'>Upload Image</h3>").
			Body(ui.Div("flex flex-col gap-4")(
				ui.ErrorForm(err, nil),

				ui.Div("flex flex-col gap-2")(
					ui.Div("text-sm font-semibold text-gray-700 dark:text-gray-300")("Result"),
					ui.Div("p-3 bg-gray-50 dark:bg-gray-800 rounded border border-gray-200 dark:border-gray-700")(
						result,
					),
				),

				form.Render(),

				// Title field
				form.Text("Title", data).Required().Render("Title"),

				// Description field
				form.Area("Description", data).Rows(3).Render("Description"),

				// Image upload component with built-in preview (combined File + ImagePreview)
				form.ImageUpload("image").
					Zone("", "Click to take or upload").
					ZoneIcon(ui.Icon("fa fa-image fa-3x")).
					MaxSize("320px").
					Required().
					Render("Image"),

				ui.Div("text-xs text-gray-500 mt-1")(
					"Accepted formats: JPG, PNG, GIF, WebP. Max size: 5MB",
				),

				// Submit button
				ui.Div("flex justify-end")(
					form.Button().Color(ui.Blue).Submit().Render("Upload Image"),
				),
			)).Render(),
	)
}

func ImageUploadContent(ctx *ui.Context) string {
	data := imageUploadData{
		Title:       "Test",
		Description: "Test description",
	}

	return renderImageUploadForm(ctx, &data, nil)
}
