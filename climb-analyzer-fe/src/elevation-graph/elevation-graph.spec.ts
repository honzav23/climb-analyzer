import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ElevationGraph } from './elevation-graph';

describe('ElevationGraph', () => {
  let component: ElevationGraph;
  let fixture: ComponentFixture<ElevationGraph>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ElevationGraph]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ElevationGraph);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
