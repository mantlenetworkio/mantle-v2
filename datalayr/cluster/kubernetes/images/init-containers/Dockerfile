FROM busybox
COPY ./data /template
CMD cp -r /template /data/ \
    && echo "Successfully copied data directory to /data" \
    || echo "Error copied data directory to /data"
