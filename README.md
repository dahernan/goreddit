What is this?
-------------

Cleverbaboon TV is a experiment using Reddit API to pull videos.
Streaming new videos to the UI, using SSE (Server Side Events)

Why?
-------------
I want to do an example in transforming a polling API (Reddit Listing API) into a Streaming API, and I like videos

Reusable parts
-------------

api/reddit - Implements listing using Reddit Rest API
oauth - Implements Application-only OAuth https://github.com/reddit/reddit/wiki/OAuth2
streaming - Transform a polling API, to a streaming API based in Go channels
sse - Minimal implementation of a http broker for Server Side Events

