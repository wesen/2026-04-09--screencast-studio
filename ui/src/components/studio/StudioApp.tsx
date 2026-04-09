import React, { useEffect, useState } from 'react';
import { useAppDispatch, useAppSelector } from '@/app/hooks';
import {
  selectSources,
  selectArmedSources,
  addSource,
  removeSource,
  updateSource,
  setFormat,
  setFps,
  setQuality,
  setAudio,
  setMultiTrack,
  setMicInput,
  setGain,
  type SourceType,
} from '@/features/studio-draft/studioDraftSlice';
import {
  selectSession,
  setElapsed,
  setMicLevel,
} from '@/features/session/sessionSlice';
import { MenuBar, SourceGrid, OutputPanel, MicPanel, StatusPanel } from './index';

export const StudioApp: React.FC = () => {
  const dispatch = useAppDispatch();
  const sources = useAppSelector(selectSources);
  const armedSources = useAppSelector(selectArmedSources);
  const session = useAppSelector(selectSession);

  const [isPaused, setIsPaused] = useState(false);
  const [elapsed, setElapsedState] = useState(0);
  const [diskPercent, setDiskPercent] = useState(8); // Simulated disk usage

  const isRecording = session.active && session.state === 'running';

  // Simulate elapsed time
  useEffect(() => {
    if (!isRecording || isPaused) return;

    const id = setInterval(() => {
      setElapsedState((prev) => prev + 1);
      dispatch(setElapsed(elapsed + 1));
    }, 1000);

    return () => clearInterval(id);
  }, [isRecording, isPaused, elapsed, dispatch]);

  // Simulate disk usage growth during recording
  useEffect(() => {
    if (!isRecording) {
      setDiskPercent(8);
      return;
    }

    const id = setInterval(() => {
      setDiskPercent((prev) => Math.min(95, prev + 0.2 * armedSources.length));
    }, 1000);

    return () => clearInterval(id);
  }, [isRecording, armedSources.length]);

  // Simulate mic level during recording
  useEffect(() => {
    if (!isRecording || isPaused) {
      dispatch(setMicLevel(0.12));
      return;
    }

    const id = setInterval(() => {
      const level = 0.15 + Math.random() * 0.6;
      dispatch(setMicLevel(level));
    }, 110);

    return () => clearInterval(id);
  }, [isRecording, isPaused, dispatch]);

  // Get output settings from draft
  const outputSettings = useAppSelector((state) => ({
    format: state.studioDraft.format,
    fps: state.studioDraft.fps,
    quality: state.studioDraft.quality,
    audio: state.studioDraft.audio,
    multiTrack: state.studioDraft.multiTrack,
  }));

  const micSettings = useAppSelector((state) => ({
    micInput: state.studioDraft.micInput,
    gain: state.studioDraft.gain,
    micLevel: state.studioDraft.micLevel,
  }));

  const handleToggleRecording = () => {
    if (isRecording) {
      // TODO: Call stopRecording API
      setElapsedState(0);
      setIsPaused(false);
    } else {
      // TODO: Call startRecording API
    }
  };

  const handleTogglePause = () => {
    setIsPaused(!isPaused);
  };

  const handleAddSource = (kind: SourceType) => {
    dispatch(addSource(kind));
  };

  const handleRemoveSource = (id: number) => {
    dispatch(removeSource(id));
  };

  const handleToggleArmed = (id: number) => {
    const source = sources.find((s) => s.id === id);
    if (source) {
      dispatch(updateSource({ id, patch: { armed: !source.armed } }));
    }
  };

  const handleToggleSolo = (id: number) => {
    const source = sources.find((s) => s.id === id);
    if (source) {
      dispatch(updateSource({ id, patch: { solo: !source.solo } }));
    }
  };

  const handleChangeScene = (id: number, scene: string) => {
    dispatch(updateSource({ id, patch: { scene } }));
  };

  return (
    <>
      <MenuBar
        armedCount={armedSources.length}
        isRecording={isRecording}
        isPaused={isPaused}
      />

      <div className="studio-main">
        <SourceGrid
          sources={sources}
          isRecording={isRecording}
          onRemove={handleRemoveSource}
          onToggleArmed={handleToggleArmed}
          onToggleSolo={handleToggleSolo}
          onChangeScene={handleChangeScene}
          onAdd={handleAddSource}
        />

        <div className="studio-content-row">
          <OutputPanel
            format={outputSettings.format}
            fps={outputSettings.fps}
            quality={outputSettings.quality}
            audio={outputSettings.audio}
            multiTrack={outputSettings.multiTrack}
            isRecording={isRecording}
            isPaused={isPaused}
            elapsed={elapsed}
            armedCount={armedSources.length}
            onFormatChange={(f) => dispatch(setFormat(f))}
            onFpsChange={(f) => dispatch(setFps(f))}
            onQualityChange={(q) => dispatch(setQuality(q))}
            onAudioChange={(a) => dispatch(setAudio(a))}
            onMultiTrackChange={(m) => dispatch(setMultiTrack(m))}
            onToggleRecording={handleToggleRecording}
            onTogglePause={handleTogglePause}
          />

          <div className="studio-panel-stack">
            <MicPanel
              micLevel={micSettings.micLevel}
              micInput={micSettings.micInput}
              gain={micSettings.gain}
              isRecording={isRecording}
              onMicInputChange={(m) => dispatch(setMicInput(m))}
              onGainChange={(g) => dispatch(setGain(g))}
            />

            <StatusPanel
              diskPercent={diskPercent}
              isRecording={isRecording}
              isPaused={isPaused}
              armedSources={armedSources}
            />
          </div>
        </div>
      </div>
    </>
  );
};
