import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ClimbGraph } from './climb-graph';

describe('ClimbGraph', () => {
  let component: ClimbGraph;
  let fixture: ComponentFixture<ClimbGraph>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ClimbGraph]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ClimbGraph);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
