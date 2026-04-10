export type StudioSourceKind = 'Display' | 'Window' | 'Region' | 'Camera';

export interface StudioSource {
  id: string;
  sourceId: string;
  kind: StudioSourceKind;
  scene: string;
  armed: boolean;
  solo: boolean;
  label: string;
  detail?: string;
  previewId?: string;
  previewState?: string;
  previewReason?: string;
  previewUrl?: string;
}

export const STUDIO_SOURCE_SCENES: Record<StudioSourceKind, string[]> = {
  Display: ['Desktop 1', 'Desktop 2'],
  Window: ['Finder', 'Terminal', 'Browser', 'Code Editor'],
  Region: ['Top Half', 'Bottom Half', 'Custom Region'],
  Camera: ['Built-in', 'USB Camera', 'FaceTime HD'],
};
