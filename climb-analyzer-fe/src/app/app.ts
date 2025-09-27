import { Component } from '@angular/core';
import { FileInput } from '../file-input/file-input';
import { TripSummaryView } from '../trip-summary-view/trip-summary-view';
import {HttpClient} from '@angular/common/http';
import { Observable } from 'rxjs';
import type { AnalysisResponse } from '../types/AnalysisResponse';
import {AsyncPipe} from '@angular/common';

@Component({
  selector: 'app-root',
  imports: [FileInput, AsyncPipe, TripSummaryView],
  templateUrl: './app.html',
  styleUrl: './app.css'
})
export class App {
  analysisResponse$: Observable<AnalysisResponse> | null = null;
  constructor(private http: HttpClient) {}

  analyzeClimbs(formData: FormData) {
    this.analysisResponse$ = this.http.post<AnalysisResponse>("http://localhost:8080/analyze", formData)
  }
}
