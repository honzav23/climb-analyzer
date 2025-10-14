import {Component, input, SimpleChanges} from '@angular/core';
import { CardModule } from "primeng/card";
import { ButtonModule } from "primeng/button";
import { ClimbGraph } from "../climb-graph/climb-graph";
import type {Climb} from "../types/AnalysisResponse";
import { OnChanges } from "@angular/core";

@Component({
  selector: 'climbs-list',
  imports: [CardModule, ButtonModule, ClimbGraph],
  templateUrl: './climbs-list.html',
  styleUrl: './climbs-list.css'
})
export class ClimbsList implements OnChanges {
    climbs = input<Climb[]>([])
    selectedClimbGraphs: boolean[] = []

    ngOnChanges(changes: SimpleChanges) {
        this.selectedClimbGraphs = Array(this.climbs().length).fill(false)
    }

    selectClimb(index: number) {
        this.selectedClimbGraphs[index] = !this.selectedClimbGraphs[index];
    }
}
