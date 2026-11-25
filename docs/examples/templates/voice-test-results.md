# Live voice test results

## 2026-11-25: listener and recorder rejoin

- **Build:** Go 1.22, repository workspace based on commit `cf45f2f` with
  then-unreleased local MLS and REST changes. Both bots used the local DAVE and MLS
  modules; this was not a test of the published `v0.18.0-walk` modules.
- **Participants:** one music bot, one recorder bot, and one consenting listener
  in a test voice channel. Both bots negotiated DAVE protocol version 1.
- **Recording:** the recorder announced the test in the channel chat and was
  configured to save only the music bot's audio. The output was 48 kHz stereo
  PCM WAV. The listener's speech was not written to the recording.
- **Playback setup:** direct playback initially failed with FFmpeg exit status
  8. The installed yt-dlp version received HTTP 403 for the media. A newer
  yt-dlp version was installed in a private temporary directory; it downloaded
  the requested track, and playback from that local file succeeded. A separate
  short FFmpeg check using the newer version's direct URL also succeeded.

### Results

1. **Baseline:** the listener confirmed they could hear the music. In an
   eight-second observation, the recorder wrote 6.48 seconds of non-silent
   audio, with RMS approximately -10.7 dBFS.
2. **Listener leaves and rejoins:** the listener confirmed hearing the song
   after returning. A fresh eight-second observation recorded 6.64 seconds of
   non-silent audio, with RMS approximately -11.5 dBFS.
3. **Recorder leaves and rejoins:** the recorder reconnected and resumed
   receiving audio. The listener confirmed that music remained audible after
   the recorder rejoined.
4. **Packet diagnosis:** in a later 36-second stable interval after restarting
   the music bot, 1,799 UDP sends returned success; 1,381 matching RTP sequence
   numbers appeared at the recorder's instrumented receive point. The other
   418 (23.2%) were not observed there. There were no DAVE or transport
   authentication errors in that stable interval. During an earlier recorder
   transition, 26 DAVE authentication failures occurred within about 50 ms
   before the stream stabilized. These observations do not identify whether
   the missing packets were lost on the network, at the Discord relay, or
   before the instrumented point in the receiver.

The rejoin and non-silent recording checks passed for this run. **Recording
continuity did not pass:** the audio duration and packet traces show missing
data. Repeat the packet comparison on another network or voice region before
attributing the loss. This single run does not establish behavior for all
membership transitions or provide a cryptographic security audit.
