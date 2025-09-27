import {Component, output} from '@angular/core';
import { ButtonModule } from 'primeng/button';
import {FormsModule} from '@angular/forms';

@Component({
  selector: 'file-input',
  imports: [
    ButtonModule,
    FormsModule
  ],
  templateUrl: './file-input.html',
  styleUrl: './file-input.css'
})
export class FileInput {
  file: File | null = null;
  isDragOver = false;

  formSubmit = output<FormData>()

  onDragOver(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragOver = true;
  }

  onDragLeave(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragOver = false;
  }

  onDrop(event: DragEvent): void {
    event.preventDefault();
    event.stopPropagation();
    this.isDragOver = false;

    if (event.dataTransfer?.files && event.dataTransfer.files.length > 0) {
      this.file = event.dataTransfer.files[0];
    }
  }

  onFileSelected(event: Event): void {
    const inputElement = event.target as HTMLInputElement;
    if (inputElement.files && inputElement.files.length > 0) {
      this.file = inputElement.files[0];
    }
  }

  clearFile(): void {
    this.file = null;
    const fileInput = document.querySelector('#file-input') as HTMLInputElement;
    if (fileInput) {
      fileInput.value = '';
    }
  }

  submitForm() {
    const formData = new FormData();
    if (this.file) {
      formData.append('file', this.file);
      this.formSubmit.emit(formData);
    }
  }
}
