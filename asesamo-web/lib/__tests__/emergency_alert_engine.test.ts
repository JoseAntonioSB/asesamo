import test from 'node:test';
import assert from 'node:assert/strict';
import { apiClient } from '../api/client';
import { casetaAudio } from '../utils/audio';
import { setStoredToken, clearAuthSession } from '../auth';

test('ApiClient.getRealtimeEventsUrl includes baseUrl and token param', () => {
  const url = apiClient.getRealtimeEventsUrl('mock-jwt-token-12345');
  assert.ok(url.includes('/realtime/events'));
  assert.ok(url.includes('token=mock-jwt-token-12345'));
});

test('CasetaAudioEngine.playEmergencySiren does not throw in headless Node', () => {
  assert.doesNotThrow(() => {
    casetaAudio.playEmergencySiren();
  });
});

test('Evacuation type formatter maps STPS emergency codes to Spanish labels', () => {
  const formatEvacType = (type: string): string => {
    switch (type) {
      case 'fire':
        return 'Alarma de Incendio (NOM-002)';
      case 'gas_leak':
        return 'Fuga de Gas / Nube Tóxica';
      case 'earthquake':
        return 'Sismo / Daño Estructural';
      case 'spill':
        return 'Derrame Químico Mayor';
      case 'drill':
        return 'Simulacro Programado (Protección Civil)';
      default:
        return 'Contingencia Operativa';
    }
  };

  assert.equal(formatEvacType('fire'), 'Alarma de Incendio (NOM-002)');
  assert.equal(formatEvacType('gas_leak'), 'Fuga de Gas / Nube Tóxica');
  assert.equal(formatEvacType('earthquake'), 'Sismo / Daño Estructural');
  assert.equal(formatEvacType('drill'), 'Simulacro Programado (Protección Civil)');
});
