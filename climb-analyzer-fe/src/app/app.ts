import { Component } from '@angular/core';
import { FileInput } from '../file-input/file-input';
import { TripSummaryView } from '../trip-summary-view/trip-summary-view';
import { ClimbsList } from "../climbs-list/climbs-list";
import {HttpClient} from '@angular/common/http';
import { Observable } from 'rxjs';
import { catchError } from 'rxjs/operators';
import type { AnalysisResponse } from '../types/AnalysisResponse';
import {AsyncPipe} from '@angular/common';
import { TabsModule } from "primeng/tabs";
import { Toast } from "primeng/toast";
import { MessageService } from "primeng/api";
import { of } from 'rxjs';

@Component({
    selector: 'app-root',
    imports: [FileInput, AsyncPipe, TripSummaryView, ClimbsList, TabsModule, Toast],
    templateUrl: './app.html',
    styleUrl: './app.css',
    providers: [MessageService]
})
export class App {
    analysisResponse$: Observable<AnalysisResponse | null> | null = null;
    constructor(private http: HttpClient, private messageService: MessageService) { }

    analyzeClimbs(formData: FormData) {
      this.analysisResponse$ = this.http.post<AnalysisResponse>("http://localhost:8080/analyze", formData).pipe(
          catchError((err) => {
              this.messageService.add({ severity: 'error', summary: 'Error', detail: err.error.error });
              return of(null);
          })
      )
    }
}
