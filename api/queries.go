package api

// GraphQL queries for the banned.video API

const GetAllChannelsQuery = `
query {
  getAllChannels {
    _id
    title
    summary
    textInfo
    avatar
    coverImage
    isLive
    totalVideos
    totalVideoViews
    totalLikes
    showInfo {
      times
      phone
      __typename
    }
    links {
      website
      facebook
      twitter
      gab
      minds
      telegram
      subscribeStar
      __typename
    }
    __typename
  }
}
`

const GetChannelWithVideosQuery = `
query GetChannelVideos($id: String!, $limit: Float, $offset: Float) {
  getChannel(id: $id) {
    _id
    title
    summary
    textInfo
    avatar
    coverImage
    isLive
    showInfo {
      times
      phone
      __typename
    }
    links {
      website
      facebook
      twitter
      gab
      minds
      telegram
      subscribeStar
      __typename
    }
    videos(limit: $limit, offset: $offset) {
      _id
      title
      summary
      largeImage
      videoDuration
      createdAt
      directUrl
      playCount
      likeCount
      angerCount
      embedUrl
      published
    }
    __typename
  }
}
`

const GetVideoQuery = `
query GetVideo($id: String!) {
  getVideo(id: $id) {
    _id
    title
    summary
    largeImage
    videoDuration
    createdAt
    directUrl
    playCount
    likeCount
    angerCount
    embedUrl
    published
    __typename
  }
}
`

const GetChannelsQuery = `
query GetChannels($ids: [String!]!) {
  getChannels(ids: $ids) {
    _id
    title
    summary
    textInfo
    avatar
    coverImage
    isLive
    totalVideos
    totalVideoViews
    totalLikes
    __typename
  }
}
`

const GetChannelQuery = `
query GetChannel($id: String!) {
  getChannel(id: $id) {
    _id
    title
    summary
    textInfo
    avatar
    coverImage
    isLive
    totalVideos
    totalVideoViews
    totalLikes
    showInfo {
      times
      phone
      __typename
    }
    links {
      website
      facebook
      twitter
      gab
      minds
      telegram
      subscribeStar
      __typename
    }
    __typename
  }
}
`

const GetChannelHotVideosQuery = `
query GetChannelHotVideos($id: String!, $limit: Float, $offset: Float) {
  getChannel(id: $id) {
    videos(limit: $limit, offset: $offset, orderBy: "likeCount") {
      _id
      title
      summary
      largeImage
      videoDuration
      createdAt
      directUrl
      playCount
      likeCount
      angerCount
      embedUrl
      published
    }
  }
}
`

const GetVideosQuery = `
query GetVideos($ids: [String!]!) {
  getVideos(ids: $ids) {
    _id
    title
    summary
    largeImage
    videoDuration
    createdAt
    directUrl
    playCount
    likeCount
    angerCount
    embedUrl
    published
    __typename
  }
}
`

const GetHotVideosQuery = `
query GetHotVideos($limit: Float, $offset: Float) {
  getHotVideos(limit: $limit, offset: $offset) {
    _id
    title
    summary
    largeImage
    videoDuration
    createdAt
    directUrl
    playCount
    likeCount
    angerCount
    embedUrl
    published
    __typename
  }
}
`

const GetNewVideosQuery = `
query GetNewVideos($limit: Float, $offset: Float) {
  getNewVideos(limit: $limit, offset: $offset) {
    _id
    title
    summary
    largeImage
    videoDuration
    createdAt
    directUrl
    playCount
    likeCount
    angerCount
    embedUrl
    published
    __typename
  }
}
`
